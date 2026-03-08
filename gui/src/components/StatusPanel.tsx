interface StatusFile { path: string; status: 'A' | 'M' | 'D' | 'U'; }

interface StatusPanelProps {
    files: StatusFile[];
    branch: string;
    loading: boolean;
}

export default function StatusPanel({ files, branch, loading }: StatusPanelProps) {
    if (loading) {
        return (
            <div className="empty-state">
                <div className="loading-spinner" />
                <div className="empty-state-text">Loading status...</div>
            </div>
        );
    }

    const staged = files.filter(f => f.status === 'A' || f.status === 'M' || f.status === 'D');
    const untracked = files.filter(f => f.status === 'U');

    if (files.length === 0) {
        return (
            <div className="empty-state animate-in">
                <div className="empty-state-icon">✅</div>
                <div className="empty-state-text">Working tree clean</div>
                <div className="empty-state-sub">
                    On branch <strong style={{ color: '#58a6ff' }}>{branch}</strong> — nothing to commit
                </div>
            </div>
        );
    }

    const statusLabels: Record<string, string> = {
        A: 'Added',
        M: 'Modified',
        D: 'Deleted',
        U: 'Untracked',
    };

    return (
        <div className="animate-in">
            <div style={{ marginBottom: 8, fontSize: 13, color: '#8b949e' }}>
                On branch <strong style={{ color: '#58a6ff' }}>{branch}</strong>
            </div>

            {staged.length > 0 && (
                <div className="status-section">
                    <div className="status-section-title">Changes to be committed</div>
                    {staged.map((file) => (
                        <div key={file.path} className="status-file">
                            <span className={`status-badge ${file.status}`}>{file.status}</span>
                            <span style={{ color: statusColor(file.status) }}>{file.path}</span>
                            <span style={{ marginLeft: 'auto', fontSize: 11, color: '#6e7681' }}>
                                {statusLabels[file.status]}
                            </span>
                        </div>
                    ))}
                </div>
            )}

            {untracked.length > 0 && (
                <div className="status-section">
                    <div className="status-section-title">Untracked files</div>
                    {untracked.map((file) => (
                        <div key={file.path} className="status-file">
                            <span className="status-badge" style={{ background: 'rgba(139, 148, 158, 0.1)', color: '#8b949e', border: '1px solid rgba(139, 148, 158, 0.3)' }}>?</span>
                            <span style={{ color: '#8b949e' }}>{file.path}</span>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}

function statusColor(status: string): string {
    switch (status) {
        case 'A': return '#3fb950';
        case 'M': return '#d29922';
        case 'D': return '#f85149';
        default: return '#8b949e';
    }
}
