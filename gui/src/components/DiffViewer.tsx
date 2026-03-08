interface DiffLine { type: 'add' | 'del' | 'ctx'; content: string; }
interface DiffResult { filePath: string; fileType: string; summary: string; lines: DiffLine[]; isCAD: boolean; additions: number; deletions: number; }
import ThreeViewer from './ThreeViewer';


interface DiffViewerProps {
    diffs: DiffResult[];
    loading: boolean;
}

export default function DiffViewer({ diffs, loading }: DiffViewerProps) {
    if (loading) {
        return (
            <div className="empty-state">
                <div className="loading-spinner" />
                <div className="empty-state-text">Computing diff...</div>
            </div>
        );
    }

    if (diffs.length === 0) {
        return (
            <div className="empty-state">
                <div className="empty-state-icon">📐</div>
                <div className="empty-state-text">No changes detected</div>
                <div className="empty-state-sub">Modify files to see diffs here</div>
            </div>
        );
    }

    const hasCAD = diffs.some(d => d.isCAD);

    return (
        <div className="animate-in">
            {hasCAD && (
                <div className="viewer-3d-container" style={{ marginBottom: 20, height: 350 }}>
                    <ThreeViewer label="old" color="#f85149" />
                    <ThreeViewer label="new" color="#58a6ff" />
                </div>
            )}

            <div className="diff-container">
                {diffs.map((d, i) => (
                    <div key={i}>
                        <div className="diff-file-header">
                            <span className="diff-file-name">{d.filePath}</span>
                            {d.isCAD && <span className="diff-file-type">{d.fileType}</span>}
                            <div className="diff-stats">
                                {d.additions > 0 && <span className="diff-stat-add">+{d.additions}</span>}
                                {d.deletions > 0 && <span className="diff-stat-del">-{d.deletions}</span>}
                            </div>
                        </div>
                        <div className="diff-body">
                            {d.lines.map((line, j) => (
                                <div key={j} className={`diff-line ${line.type}`}>
                                    <span className="diff-line-prefix">
                                        {line.type === 'add' ? '+' : line.type === 'del' ? '-' : ' '}
                                    </span>
                                    <span className="diff-line-content">{line.content}</span>
                                </div>
                            ))}
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
}
