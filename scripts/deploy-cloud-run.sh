#!/bin/bash
set -euo pipefail

usage() {
    cat <<'EOF'
Usage:
  scripts/deploy-cloud-run.sh <environment>

Arguments:
  environment   staging | production

Required environment variables:
  GCP_PROJECT_ID

Optional environment variables:
  GCP_REGION                  Default: asia-northeast1
  ARTIFACT_REGISTRY_REPO      Default: chronome
  CLOUD_SQL_INSTANCE          Default: chronome-db
  CLOUD_RUN_SERVICE_ACCOUNT   Default: chronome-runtime@<project>.iam.gserviceaccount.com
  BACKEND_SERVICE_NAME        Default: chronome-backend-<environment>
  FRONTEND_SERVICE_NAME       Default: chronome-frontend-<environment>
  FRONTEND_API_URL            Existing backend URL. If empty, derived after backend deploy.
  BACKEND_MIN_INSTANCES       Default: 0
  BACKEND_MAX_INSTANCES       Default: 3
  FRONTEND_MIN_INSTANCES      Default: 0
  FRONTEND_MAX_INSTANCES      Default: 3

Required Secret Manager secrets:
  chronome-db-dsn
  chronome-session-secret
EOF
}

if [ "${1:-}" = "" ] || [ "${1:-}" = "--help" ] || [ "${1:-}" = "-h" ]; then
    usage
    exit 0
fi

ENVIRONMENT="$1"
if [ "${ENVIRONMENT}" != "staging" ] && [ "${ENVIRONMENT}" != "production" ]; then
    echo "environment must be 'staging' or 'production'" >&2
    exit 1
fi

if [ -z "${GCP_PROJECT_ID:-}" ]; then
    echo "GCP_PROJECT_ID is required" >&2
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

GCP_REGION="${GCP_REGION:-asia-northeast1}"
ARTIFACT_REGISTRY_REPO="${ARTIFACT_REGISTRY_REPO:-chronome}"
CLOUD_SQL_INSTANCE="${CLOUD_SQL_INSTANCE:-chronome-db}"
CLOUD_RUN_SERVICE_ACCOUNT="${CLOUD_RUN_SERVICE_ACCOUNT:-chronome-runtime@${GCP_PROJECT_ID}.iam.gserviceaccount.com}"
BACKEND_SERVICE_NAME="${BACKEND_SERVICE_NAME:-chronome-backend-${ENVIRONMENT}}"
FRONTEND_SERVICE_NAME="${FRONTEND_SERVICE_NAME:-chronome-frontend-${ENVIRONMENT}}"
BACKEND_MIN_INSTANCES="${BACKEND_MIN_INSTANCES:-0}"
BACKEND_MAX_INSTANCES="${BACKEND_MAX_INSTANCES:-3}"
FRONTEND_MIN_INSTANCES="${FRONTEND_MIN_INSTANCES:-0}"
FRONTEND_MAX_INSTANCES="${FRONTEND_MAX_INSTANCES:-3}"
IMAGE_PREFIX="${GCP_REGION}-docker.pkg.dev/${GCP_PROJECT_ID}/${ARTIFACT_REGISTRY_REPO}"
BACKEND_IMAGE="${IMAGE_PREFIX}/${BACKEND_SERVICE_NAME}"
FRONTEND_IMAGE="${IMAGE_PREFIX}/${FRONTEND_SERVICE_NAME}"
IMAGE_TAG="${ENVIRONMENT}-$(date +%Y%m%d-%H%M%S)"
CONNECTION_NAME="${GCP_PROJECT_ID}:${GCP_REGION}:${CLOUD_SQL_INSTANCE}"

echo "Using project: ${GCP_PROJECT_ID}"
echo "Using region: ${GCP_REGION}"
echo "Deploying environment: ${ENVIRONMENT}"
echo "Using runtime service account: ${CLOUD_RUN_SERVICE_ACCOUNT}"

gcloud config set project "${GCP_PROJECT_ID}" >/dev/null

echo "Building backend image with Cloud Build..."
gcloud builds submit "${REPO_ROOT}/backend" \
    --config "${REPO_ROOT}/cloudbuild/backend.yaml" \
    --substitutions "_IMAGE=${BACKEND_IMAGE}:${IMAGE_TAG}"

echo "Deploying backend to Cloud Run..."
gcloud run deploy "${BACKEND_SERVICE_NAME}" \
    --image "${BACKEND_IMAGE}:${IMAGE_TAG}" \
    --region "${GCP_REGION}" \
    --platform managed \
    --allow-unauthenticated \
    --service-account "${CLOUD_RUN_SERVICE_ACCOUNT}" \
    --add-cloudsql-instances "${CONNECTION_NAME}" \
    --set-env-vars "APP_ENV=${ENVIRONMENT},SERVER_ADDRESS=:8080,DB_DRIVER=postgres,SESSION_COOKIE_SECURE=true,SESSION_COOKIE_SAMESITE=none,ALLOWED_ORIGIN=http://localhost:3000" \
    --set-secrets "DB_DSN=chronome-db-dsn:latest,SESSION_SECRET=chronome-session-secret:latest" \
    --min-instances "${BACKEND_MIN_INSTANCES}" \
    --max-instances "${BACKEND_MAX_INSTANCES}"

BACKEND_URL="${FRONTEND_API_URL:-$(gcloud run services describe "${BACKEND_SERVICE_NAME}" --region "${GCP_REGION}" --format='value(status.url)')}"

echo "Building frontend image with API URL: ${BACKEND_URL}"
gcloud builds submit "${REPO_ROOT}/frontend" \
    --config "${REPO_ROOT}/cloudbuild/frontend.yaml" \
    --substitutions "_IMAGE=${FRONTEND_IMAGE}:${IMAGE_TAG},_VITE_API_BASE_URL=${BACKEND_URL}"

echo "Deploying frontend to Cloud Run..."
gcloud run deploy "${FRONTEND_SERVICE_NAME}" \
    --image "${FRONTEND_IMAGE}:${IMAGE_TAG}" \
    --region "${GCP_REGION}" \
    --platform managed \
    --allow-unauthenticated \
    --service-account "${CLOUD_RUN_SERVICE_ACCOUNT}" \
    --port 8080 \
    --min-instances "${FRONTEND_MIN_INSTANCES}" \
    --max-instances "${FRONTEND_MAX_INSTANCES}"

FRONTEND_URL="$(gcloud run services describe "${FRONTEND_SERVICE_NAME}" --region "${GCP_REGION}" --format='value(status.url)')"

echo "Updating backend CORS origin to: ${FRONTEND_URL}"
gcloud run services update "${BACKEND_SERVICE_NAME}" \
    --region "${GCP_REGION}" \
    --set-env-vars "APP_ENV=${ENVIRONMENT},SERVER_ADDRESS=:8080,DB_DRIVER=postgres,SESSION_COOKIE_SECURE=true,SESSION_COOKIE_SAMESITE=none,ALLOWED_ORIGIN=${FRONTEND_URL}"

echo ""
echo "Deployment complete"
echo "Backend URL : ${BACKEND_URL}"
echo "Frontend URL: ${FRONTEND_URL}"
