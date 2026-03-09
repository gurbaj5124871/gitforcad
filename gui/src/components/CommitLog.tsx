interface CommitInfo { hash: string; message: string; author: string; date: string; branch?: string; }

interface CommitLogProps {
    commits: CommitInfo[];
    loading: boolean;
}

export default function CommitLog({ commits, loading }: CommitLogProps) {
    if (loading) {
        return (
            <div className="empty-state">
                <div className="loading-spinner" />
                <div className="empty-state-text">Loading history...</div>
            </div>
        );
    }

    if (commits.length === 0) {
        return (
            <div className="empty-state">
                <div className="empty-state-icon">📋</div>
                <div className="empty-state-text">No commits yet</div>
                <div className="empty-state-sub">Make your first commit to see history</div>
            </div>
        );
    }

    return (
        <div className="commit-list animate-in">
            {commits.map((commit, _i) => (
                <div key={commit.hash} className="commit-item">
                    <div className="commit-dot" />
                    <div className="commit-info">
                        <div className="commit-message">
                            {commit.message}
                            {commit.branch && (
                                <span style={{
                                    fontSize: 11,
                                    fontWeight: 600,
                                    marginLeft: 8,
                                    padding: '2px 8px',
                                    borderRadius: 4,
                                    background: 'rgba(88, 166, 255, 0.1)',
                                    color: '#58a6ff',
                                    border: '1px solid rgba(88, 166, 255, 0.3)',
                                }}>
                                    {commit.branch}
                                </span>
                            )}
                        </div>
                        <div className="commit-meta">
                            <span className="commit-hash">{commit.hash}</span>
                            <span>{commit.author}</span>
                            <span>{formatDate(commit.date)}</span>
                        </div>
                    </div>
                </div>
            ))}
        </div>
    );
}

function formatDate(dateStr: string): string {
    if (!dateStr) return '';
    try {
        const date = new Date(dateStr);
        const now = new Date();
        const diff = now.getTime() - date.getTime();
        const mins = Math.floor(diff / 60000);
        const hours = Math.floor(mins / 60);
        const days = Math.floor(hours / 24);

        if (mins < 1) return 'just now';
        if (mins < 60) return `${mins}m ago`;
        if (hours < 24) return `${hours}h ago`;
        if (days < 7) return `${days}d ago`;
        return date.toLocaleDateString();
    } catch {
        return dateStr;
    }
}
