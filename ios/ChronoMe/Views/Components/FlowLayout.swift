import SwiftUI

/// タグを折り返し表示するカスタムLayout
/// SwiftUIのLayout protocolを実装
struct FlowLayout: Layout {
    var spacing: CGFloat = 8
    var lineSpacing: CGFloat = 8

    func sizeThatFits(proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) -> CGSize {
        let rows = computeRows(proposal: proposal, subviews: subviews)
        let width = proposal.width ?? 0
        let height = rows.reduce(0) { $0 + $1.height } + CGFloat(max(rows.count - 1, 0)) * lineSpacing
        return CGSize(width: width, height: max(height, 0))
    }

    func placeSubviews(in bounds: CGRect, proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) {
        let rows = computeRows(proposal: proposal, subviews: subviews)
        var yOffset = bounds.minY

        for row in rows {
            var xOffset = bounds.minX

            for index in row.indices {
                let view = subviews[index]
                let size = view.sizeThatFits(.unspecified)
                view.place(at: CGPoint(x: xOffset, y: yOffset), proposal: .unspecified)
                xOffset += size.width + spacing
            }

            yOffset += row.height + lineSpacing
        }
    }

    private func computeRows(proposal: ProposedViewSize, subviews: Subviews) -> [Row] {
        var rows: [Row] = []
        var currentRow = Row(indices: [], height: 0)
        var currentWidth: CGFloat = 0
        let maxWidth = proposal.width ?? .infinity

        for (index, subview) in subviews.enumerated() {
            let size = subview.sizeThatFits(.unspecified)

            if currentWidth + size.width > maxWidth && !currentRow.indices.isEmpty {
                // 現在の行が最大幅を超える場合、新しい行を開始
                rows.append(currentRow)
                currentRow = Row(indices: [index], height: size.height)
                currentWidth = size.width + spacing
            } else {
                // 現在の行に追加
                currentRow.indices.append(index)
                currentRow.height = max(currentRow.height, size.height)
                currentWidth += size.width + spacing
            }
        }

        if !currentRow.indices.isEmpty {
            rows.append(currentRow)
        }

        return rows
    }

    struct Row {
        var indices: [Int]
        var height: CGFloat
    }
}
