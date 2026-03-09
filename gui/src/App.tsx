import { useState, useCallback, useEffect } from 'react';
import DiffViewer from './components/DiffViewer';
import CommitLog from './components/CommitLog';
import StatusPanel from './components/StatusPanel';

// Detect if running inside Tauri
const isTauri = typeof window !== 'undefined' && !!(window as any).__TAURI_INTERNALS__;

// ── Types ──
interface DiffLine { type: 'add' | 'del' | 'ctx'; content: string; }
interface DiffResult { filePath: string; fileType: string; summary: string; lines: DiffLine[]; isCAD: boolean; additions: number; deletions: number; }
interface CommitInfo { hash: string; message: string; author: string; date: string; branch?: string; }
interface StatusFile { path: string; status: 'A' | 'M' | 'D' | 'U'; }
interface BranchInfo { name: string; active: boolean; }

// ── Mock data ──
const MOCK_BRANCHES: BranchInfo[] = [
  { name: 'main', active: true },
  { name: 'feature/chamfer-edge', active: false },
  { name: 'fix/tolerance-update', active: false },
];

const MOCK_STATUS_FILES: StatusFile[] = [
  { path: 'housing.stl', status: 'M' },
  { path: 'bracket.dxf', status: 'M' },
  { path: 'assembly.obj', status: 'A' },
  { path: 'gasket.step', status: 'A' },
];

const MOCK_COMMITS: CommitInfo[] = [
  { hash: 'a3f8b2c1e4d7', message: 'Add chamfer to housing edge', author: 'engineer', date: new Date(Date.now() - 3600000).toISOString(), branch: 'main' },
  { hash: 'e7d2c8a1f5b3', message: 'Update bracket holes 4mm → 5mm', author: 'engineer', date: new Date(Date.now() - 7200000).toISOString() },
  { hash: 'b1c4f7e2a8d5', message: 'Add gasket seal profile', author: 'designer', date: new Date(Date.now() - 86400000).toISOString() },
  { hash: 'f5a2d8c1b7e3', message: 'Initial CAD design', author: 'engineer', date: new Date(Date.now() - 172800000).toISOString() },
  { hash: 'c8e1a5f3d2b7', message: 'Project initialization', author: 'engineer', date: new Date(Date.now() - 259200000).toISOString() },
];

const MOCK_DIFFS: DiffResult[] = [
  {
    filePath: 'housing.stl', fileType: 'stl', isCAD: true, additions: 192, deletions: 44,
    summary: 'STL: 1248 → 1396 triangles (+192 -44)',
    lines: [
      { type: 'del', content: 'Triangles: 1248' },
      { type: 'add', content: 'Triangles: 1396' },
      { type: 'del', content: 'Removed 44 triangles from mesh' },
      { type: 'add', content: 'Added 192 triangles to mesh' },
      { type: 'del', content: 'Bounding box: (0, 0, 0) → (50, 30, 25)' },
      { type: 'add', content: 'Bounding box: (0, 0, 0) → (52, 30, 25)' },
      { type: 'del', content: 'Surface area: 6240.00' },
      { type: 'add', content: 'Surface area: 6580.50' },
    ],
  },
  {
    filePath: 'bracket.dxf', fileType: 'dxf', isCAD: true, additions: 6, deletions: 2,
    summary: 'DXF: 34 → 38 entities (+6 -2)',
    lines: [
      { type: 'ctx', content: 'LINE: 12 (unchanged)' },
      { type: 'del', content: 'CIRCLE: 4' },
      { type: 'add', content: 'CIRCLE: 8' },
      { type: 'del', content: 'ARC: 6' },
      { type: 'add', content: 'ARC: 6 (unchanged)' },
      { type: 'ctx', content: '--- Layers ---' },
      { type: 'add', content: 'Layer: MOUNTING_HOLES (8 entities)' },
      { type: 'del', content: 'Layer: HOLES (4 entities)' },
    ],
  },
  {
    filePath: 'assembly.obj', fileType: 'obj', isCAD: true, additions: 842, deletions: 0,
    summary: 'new OBJ: assembly.obj (842 vertices, 1560 faces)',
    lines: [
      { type: 'add', content: 'Vertices: 842' },
      { type: 'add', content: 'Faces: 1560' },
      { type: 'add', content: 'Normals: 842' },
      { type: 'add', content: 'Group added: housing' },
      { type: 'add', content: 'Group added: bracket' },
      { type: 'add', content: 'Group added: gasket' },
    ],
  },
  {
    filePath: 'floorplan.dwg', fileType: 'dwg', isCAD: true, additions: 15, deletions: 4,
    summary: 'DWG (via DXF): 45 → 56 entities (+15 -4)',
    lines: [
      { type: 'ctx', content: 'LINE: 20 (unchanged)' },
      { type: 'del', content: 'CIRCLE: 1' },
      { type: 'add', content: 'CIRCLE: 3' },
      { type: 'ctx', content: 'ARC: 12 (unchanged)' },
      { type: 'del', content: 'TEXT: 4' },
      { type: 'add', content: 'TEXT: 12' },
      { type: 'add', content: 'POLYLINE: 5' },
      { type: 'ctx', content: '--- Layers ---' },
      { type: 'add', content: 'Layer: WALLS (22 entities)' },
      { type: 'del', content: 'Layer: WALLS (18 entities)' },
      { type: 'add', content: 'Layer: PLUMBING (5 entities)' },
    ],
  },
];

type Tab = 'status' | 'diff' | 'log';

function App() {
  const [repoPath, setRepoPath] = useState(isTauri ? '/tmp/dwgtest' : '/demo/cad-project');
  const [isRepoOpen, setIsRepoOpen] = useState(true);
  const [activeTab, setActiveTab] = useState<Tab>('diff');

  const [branches, setBranches] = useState<BranchInfo[]>(isTauri ? [] : MOCK_BRANCHES);
  const [currentBranch, setCurrentBranch] = useState('main');
  const [statusFiles, setStatusFiles] = useState<StatusFile[]>(isTauri ? [] : MOCK_STATUS_FILES);
  const [commits, setCommits] = useState<CommitInfo[]>(isTauri ? [] : MOCK_COMMITS);
  const [diffs, setDiffs] = useState<DiffResult[]>(isTauri ? [] : MOCK_DIFFS);

  const [loading, setLoading] = useState(false);
  const [pathInput, setPathInput] = useState('');
  const [newBranchName, setNewBranchName] = useState('');
  const [commitMsg, setCommitMsg] = useState('');
  const [showNewBranch, setShowNewBranch] = useState(false);
  const [showCommit, setShowCommit] = useState(false);
  const [error, setError] = useState('');

  const refresh = useCallback(async () => {
    if (!isTauri || !repoPath) return;
    setLoading(true);
    setError('');
    try {
      const { getStatus, getLog, getDiff, getBranches } = await import('./services/gitforcad');
      const [statusResult, logResult, diffResult, branchResult] = await Promise.allSettled([
        getStatus(repoPath), getLog(repoPath), getDiff(repoPath), getBranches(repoPath),
      ]);
      if (statusResult.status === 'fulfilled') { setStatusFiles(statusResult.value.files); setCurrentBranch(statusResult.value.branch); }
      if (logResult.status === 'fulfilled') setCommits(logResult.value);
      if (diffResult.status === 'fulfilled') setDiffs(diffResult.value);
      if (branchResult.status === 'fulfilled') setBranches(branchResult.value);
    } catch (e: any) {
      setError(e.message || 'Failed to load');
    }
    setLoading(false);
  }, [repoPath]);

  useEffect(() => { if (isRepoOpen && isTauri) refresh(); }, [isRepoOpen, refresh]);

  const handleOpenRepo = () => { if (pathInput.trim()) { setRepoPath(pathInput.trim()); setIsRepoOpen(true); } };

  const handleInitAndOpen = async () => {
    if (!pathInput.trim()) return;
    if (isTauri) {
      try { const { initRepo } = await import('./services/gitforcad'); await initRepo(pathInput.trim()); } catch (e: any) { setError(e.message); return; }
    }
    setRepoPath(pathInput.trim());
    setIsRepoOpen(true);
  };

  const handleCheckout = async (branch: string) => {
    if (isTauri) {
      try { const { checkoutBranch } = await import('./services/gitforcad'); await checkoutBranch(repoPath, branch); await refresh(); } catch (e: any) { setError(e.message); }
    }
  };

  const handleCreateBranch = async () => {
    if (!newBranchName.trim()) return;
    if (isTauri) {
      try { const { createBranch } = await import('./services/gitforcad'); await createBranch(repoPath, newBranchName.trim()); } catch (e: any) { setError(e.message); return; }
    }
    setNewBranchName(''); setShowNewBranch(false);
    if (isTauri) await refresh();
  };

  const handleCommit = async () => {
    if (!commitMsg.trim()) return;
    if (isTauri) {
      try { const { addFiles, commit } = await import('./services/gitforcad'); await addFiles(repoPath, ['.']); await commit(repoPath, commitMsg.trim()); } catch (e: any) { setError(e.message); return; }
    }
    setCommitMsg(''); setShowCommit(false);
    if (isTauri) await refresh();
  };

  const handleMerge = async (branch: string) => {
    if (isTauri) {
      try { const { mergeBranch } = await import('./services/gitforcad'); await mergeBranch(repoPath, branch); await refresh(); } catch (e: any) { setError(e.message); }
    }
  };

  // ── Repo Selector (Tauri mode only) ──
  if (!isRepoOpen) {
    return (
      <div className="repo-selector">
        <h1><img src="/logo.png" alt="GitForCAD" style={{ height: 48, verticalAlign: 'middle', marginRight: 12 }} />GitForCAD</h1>
        <p>Version control for CAD files — DWG, STL, DXF, OBJ and more</p>
        <div className="repo-input-group">
          <input type="text" placeholder="/path/to/your/cad/project" value={pathInput}
            onChange={(e) => setPathInput(e.target.value)} onKeyDown={(e) => e.key === 'Enter' && handleOpenRepo()} />
          <button className="btn btn-primary" onClick={handleOpenRepo}>Open</button>
          <button className="btn" onClick={handleInitAndOpen}>Init & Open</button>
        </div>
        {error && <p style={{ color: '#f85149', marginTop: 8 }}>{error}</p>}
        <p style={{ fontSize: 12, color: '#6e7681', marginTop: 16 }}>Enter a path to a gitforcad repository, or initialize a new one</p>
      </div>
    );
  }

  // ── Main App ──
  return (
    <div className="app-layout">
      <div className="top-bar">
        <div className="top-bar-left">
          <span className="app-logo">GitForCAD</span>
          <span className="repo-path">{repoPath}</span>
        </div>
        <div className="top-bar-right">
          <button className="btn btn-sm" onClick={refresh}>↻ Refresh</button>
          {!showCommit ? (
            <button className="btn btn-sm btn-primary" onClick={() => setShowCommit(true)}>+ Commit</button>
          ) : (
            <div style={{ display: 'flex', gap: 6, alignItems: 'center' }}>
              <input type="text" placeholder="Commit message..." value={commitMsg}
                onChange={e => setCommitMsg(e.target.value)} onKeyDown={e => e.key === 'Enter' && handleCommit()}
                style={{
                  padding: '4px 10px', fontSize: 12, background: 'var(--bg-tertiary)', border: '1px solid var(--border)',
                  borderRadius: 6, color: 'var(--text-primary)', fontFamily: 'inherit', width: 200, outline: 'none'
                }} />
              <button className="btn btn-sm btn-primary" onClick={handleCommit}>Commit</button>
              <button className="btn btn-sm" onClick={() => setShowCommit(false)}>✕</button>
            </div>
          )}
        </div>
      </div>

      {error && (
        <div style={{
          padding: '6px 16px', background: 'var(--red-bg)', borderBottom: '1px solid var(--red-border)',
          fontSize: 13, color: 'var(--red)', display: 'flex', justifyContent: 'space-between'
        }}>
          <span>{error}</span>
          <span style={{ cursor: 'pointer' }} onClick={() => setError('')}>✕</span>
        </div>
      )}

      <div className="main-content">
        <div className="sidebar">
          <div className="sidebar-section">
            <div className="section-header">
              <span>Branches</span>
              <button className="btn btn-sm" onClick={() => setShowNewBranch(!showNewBranch)}>+</button>
            </div>
            {showNewBranch && (
              <div style={{ padding: '4px 14px 8px', display: 'flex', gap: 4 }}>
                <input type="text" placeholder="Branch name" value={newBranchName}
                  onChange={e => setNewBranchName(e.target.value)} onKeyDown={e => e.key === 'Enter' && handleCreateBranch()}
                  style={{
                    flex: 1, padding: '4px 8px', fontSize: 12, background: 'var(--bg-tertiary)',
                    border: '1px solid var(--border)', borderRadius: 4, color: 'var(--text-primary)', fontFamily: 'inherit', outline: 'none'
                  }} />
                <button className="btn btn-sm" onClick={handleCreateBranch}>✓</button>
              </div>
            )}
            <div className="section-content">
              {branches.map(b => (
                <div key={b.name} className={`branch-item ${b.active ? 'active' : ''}`}
                  onClick={() => !b.active && handleCheckout(b.name)}>
                  <span className="branch-icon">{b.active ? '◉' : '○'}</span>
                  <span>{b.name}</span>
                  {!b.active && b.name !== currentBranch && (
                    <button className="btn btn-sm" style={{ marginLeft: 'auto', fontSize: 10, padding: '2px 6px' }}
                      onClick={(e) => { e.stopPropagation(); handleMerge(b.name); }}>Merge</button>
                  )}
                </div>
              ))}
            </div>
          </div>

          <div className="sidebar-section expand">
            <div className="section-header">
              <span>Changed Files</span>
              <span style={{ fontSize: 10, color: 'var(--text-muted)' }}>{statusFiles.length}</span>
            </div>
            <div className="section-content">
              {statusFiles.map(f => (
                <div key={f.path} className="file-item">
                  <span style={{ fontSize: 12, opacity: 0.5 }}>📄</span>
                  <span style={{ fontSize: 12, overflow: 'hidden', textOverflow: 'ellipsis' }}>{f.path}</span>
                  <span className={`file-status ${f.status === 'A' ? 'added' : f.status === 'M' ? 'modified' : f.status === 'D' ? 'deleted' : ''}`}>{f.status}</span>
                </div>
              ))}
              {statusFiles.length === 0 && (
                <div style={{ padding: '12px 14px', fontSize: 12, color: 'var(--text-muted)' }}>No changes</div>
              )}
            </div>
          </div>
        </div>

        <div className="center-panel">
          <div className="panel-tabs">
            {([['status', 'Status'], ['diff', 'Diff'], ['log', 'History']] as [Tab, string][]).map(([key, label]) => (
              <div key={key} className={`panel-tab ${activeTab === key ? 'active' : ''}`} onClick={() => setActiveTab(key)}>{label}</div>
            ))}
          </div>

          <div className="panel-content">
            {activeTab === 'status' && <StatusPanel files={statusFiles} branch={currentBranch} loading={loading} />}
            {activeTab === 'diff' && <DiffViewer diffs={diffs} loading={loading} />}
            {activeTab === 'log' && <CommitLog commits={commits} loading={loading} />}
          </div>
        </div>
      </div>

      <div className="status-bar">
        <div className="status-bar-left">
          <div className="status-indicator"><div className="status-dot" /><span>branch: {currentBranch}</span></div>
          <span>{commits.length} commit(s)</span>
        </div>
        <div className="status-bar-right"><span>GitForCAD v0.1.0</span></div>
      </div>
    </div>
  );
}

export default App;
