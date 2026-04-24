#!/bin/bash
set -euo pipefail

# ChronoMe GCP Setup Script
# This script sets up the required GCP resources for deploying ChronoMe

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
echo_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
echo_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Check if required environment variables are set
check_env() {
    if [ -z "${GCP_PROJECT_ID:-}" ]; then
        echo_error "GCP_PROJECT_ID environment variable is required"
        echo "Usage: GCP_PROJECT_ID=your-project-id GITHUB_REPO=owner/repo ./setup-gcp.sh"
        exit 1
    fi
    if [ -z "${GITHUB_REPO:-}" ]; then
        echo_error "GITHUB_REPO environment variable is required (format: owner/repo)"
        exit 1
    fi
}

# Configuration
REGION="asia-northeast1"
ARTIFACT_REGISTRY_REPO="chronome"
CLOUD_SQL_INSTANCE="chronome-db"
WORKLOAD_IDENTITY_POOL="github-pool"
WORKLOAD_IDENTITY_PROVIDER="github-provider"
SERVICE_ACCOUNT_NAME="github-actions"

main() {
    check_env

    echo_info "Setting up GCP resources for project: ${GCP_PROJECT_ID}"
    echo_info "GitHub repository: ${GITHUB_REPO}"

    # Set project
    gcloud config set project "${GCP_PROJECT_ID}"

    # Enable required APIs
    echo_info "Enabling required APIs..."
    gcloud services enable \
        run.googleapis.com \
        sqladmin.googleapis.com \
        artifactregistry.googleapis.com \
        secretmanager.googleapis.com \
        iamcredentials.googleapis.com \
        cloudresourcemanager.googleapis.com

    # Create Artifact Registry repository
    echo_info "Creating Artifact Registry repository..."
    if ! gcloud artifacts repositories describe "${ARTIFACT_REGISTRY_REPO}" --location="${REGION}" &>/dev/null; then
        gcloud artifacts repositories create "${ARTIFACT_REGISTRY_REPO}" \
            --repository-format=docker \
            --location="${REGION}" \
            --description="ChronoMe Docker images"
    else
        echo_warn "Artifact Registry repository already exists"
    fi

    # Create Cloud SQL instance
    echo_info "Creating Cloud SQL instance (this may take several minutes)..."
    if ! gcloud sql instances describe "${CLOUD_SQL_INSTANCE}" &>/dev/null; then
        gcloud sql instances create "${CLOUD_SQL_INSTANCE}" \
            --database-version=POSTGRES_16 \
            --tier=db-f1-micro \
            --region="${REGION}" \
            --root-password="$(openssl rand -base64 24)" \
            --database-flags=cloudsql.iam_authentication=on

        # Create database
        gcloud sql databases create chronome --instance="${CLOUD_SQL_INSTANCE}"

        # Create user
        DB_PASSWORD=$(openssl rand -base64 24)
        gcloud sql users create chronome \
            --instance="${CLOUD_SQL_INSTANCE}" \
            --password="${DB_PASSWORD}"

        echo_info "Database password: ${DB_PASSWORD}"
        echo_warn "Please save this password securely!"
    else
        echo_warn "Cloud SQL instance already exists"
    fi

    # Get Cloud SQL connection name
    CONNECTION_NAME=$(gcloud sql instances describe "${CLOUD_SQL_INSTANCE}" --format='value(connectionName)')
    echo_info "Cloud SQL connection name: ${CONNECTION_NAME}"

    # Create secrets
    echo_info "Setting up Secret Manager secrets..."

    # Create DB_DSN secret
    if ! gcloud secrets describe chronome-db-dsn &>/dev/null; then
        echo_warn "Creating chronome-db-dsn secret..."
        echo "Please add the DB_DSN value manually:"
        echo "  gcloud secrets create chronome-db-dsn --replication-policy=automatic"
        echo "  echo -n 'host=/cloudsql/${CONNECTION_NAME} user=chronome password=YOUR_PASSWORD dbname=chronome' | gcloud secrets versions add chronome-db-dsn --data-file=-"
    fi

    # Create SESSION_SECRET
    if ! gcloud secrets describe chronome-session-secret &>/dev/null; then
        SESSION_SECRET=$(openssl rand -base64 32)
        echo -n "${SESSION_SECRET}" | gcloud secrets create chronome-session-secret \
            --replication-policy=automatic \
            --data-file=-
        echo_info "Session secret created"
    else
        echo_warn "Session secret already exists"
    fi

    # Setup Workload Identity Federation for GitHub Actions
    echo_info "Setting up Workload Identity Federation..."

    # Create service account
    if ! gcloud iam service-accounts describe "${SERVICE_ACCOUNT_NAME}@${GCP_PROJECT_ID}.iam.gserviceaccount.com" &>/dev/null; then
        gcloud iam service-accounts create "${SERVICE_ACCOUNT_NAME}" \
            --display-name="GitHub Actions Service Account"
    else
        echo_warn "Service account already exists"
    fi

    # Grant required roles to service account
    echo_info "Granting IAM roles to service account..."
    ROLES=(
        "roles/run.admin"
        "roles/artifactregistry.writer"
        "roles/secretmanager.secretAccessor"
        "roles/iam.serviceAccountUser"
    )
    for role in "${ROLES[@]}"; do
        gcloud projects add-iam-policy-binding "${GCP_PROJECT_ID}" \
            --member="serviceAccount:${SERVICE_ACCOUNT_NAME}@${GCP_PROJECT_ID}.iam.gserviceaccount.com" \
            --role="${role}" \
            --quiet
    done

    # Create Workload Identity Pool
    if ! gcloud iam workload-identity-pools describe "${WORKLOAD_IDENTITY_POOL}" --location=global &>/dev/null; then
        gcloud iam workload-identity-pools create "${WORKLOAD_IDENTITY_POOL}" \
            --location=global \
            --display-name="GitHub Actions Pool"
    else
        echo_warn "Workload Identity Pool already exists"
    fi

    # Create Workload Identity Provider
    if ! gcloud iam workload-identity-pools providers describe "${WORKLOAD_IDENTITY_PROVIDER}" \
        --location=global \
        --workload-identity-pool="${WORKLOAD_IDENTITY_POOL}" &>/dev/null; then
        gcloud iam workload-identity-pools providers create-oidc "${WORKLOAD_IDENTITY_PROVIDER}" \
            --location=global \
            --workload-identity-pool="${WORKLOAD_IDENTITY_POOL}" \
            --display-name="GitHub Provider" \
            --issuer-uri="https://token.actions.githubusercontent.com" \
            --attribute-mapping="google.subject=assertion.sub,attribute.actor=assertion.actor,attribute.repository=assertion.repository" \
            --attribute-condition="assertion.repository=='${GITHUB_REPO}'"
    else
        echo_warn "Workload Identity Provider already exists"
    fi

    # Allow GitHub Actions to impersonate service account
    gcloud iam service-accounts add-iam-policy-binding \
        "${SERVICE_ACCOUNT_NAME}@${GCP_PROJECT_ID}.iam.gserviceaccount.com" \
        --role="roles/iam.workloadIdentityUser" \
        --member="principalSet://iam.googleapis.com/projects/$(gcloud projects describe ${GCP_PROJECT_ID} --format='value(projectNumber)')/locations/global/workloadIdentityPools/${WORKLOAD_IDENTITY_POOL}/attribute.repository/${GITHUB_REPO}"

    # Get project number
    PROJECT_NUMBER=$(gcloud projects describe "${GCP_PROJECT_ID}" --format='value(projectNumber)')

    echo ""
    echo_info "Setup complete!"
    echo ""
    echo "Required GitHub Secrets:"
    echo "========================"
    echo "GCP_PROJECT_ID: ${GCP_PROJECT_ID}"
    echo "GCP_PROJECT_NUMBER: ${PROJECT_NUMBER}"
    echo ""
    echo "Required URL Secrets (set after first deploy):"
    echo "STAGING_BACKEND_URL: https://chronome-backend-staging-xxxxx-an.a.run.app"
    echo "STAGING_FRONTEND_URL: https://chronome-frontend-staging-xxxxx-an.a.run.app"
    echo "PRODUCTION_BACKEND_URL: https://chronome-backend-xxxxx-an.a.run.app"
    echo "PRODUCTION_FRONTEND_URL: https://chronome-frontend-xxxxx-an.a.run.app"
    echo ""
    echo "Cloud SQL Connection Name: ${CONNECTION_NAME}"
    echo ""
    echo_warn "Don't forget to set the chronome-db-dsn secret with the database connection string!"
}

main "$@"
