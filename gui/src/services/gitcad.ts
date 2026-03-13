// Detect if we're running inside Tauri
export const isTauri = typeof window !== 'undefined' && !!(window as any).__TAURI_INTERNALS__;

export interface DiffLine {
    type: 'add' | 'del' | 'ctx';
    content: string;
}

export interface DiffResult {
    filePath: string;
    fileType: string;
    summary: string;
    lines: DiffLine[];
    isCAD: boolean;
    additions: number;
    deletions: number;
}

export interface CommitInfo {
    hash: string;
    message: string;
    author: string;
    date: string;
    branch?: string;
}

export interface StatusFile {
    path: string;
    status: 'A' | 'M' | 'D' | 'U';
}

export interface BranchInfo {
    name: string;
    active: boolean;
}

// ── Tauri shell execution ──

async function runGitCAD(args: string[], cwd: string): Promise<string> {
    // Dynamic import to avoid loading Tauri modules in browser
    const { Command } = await import('@tauri-apps/plugin-shell');
    const cmd = Command.create('gitcad', args, { cwd });
    const output = await cmd.execute();
    if (output.code !== 0) {
        throw new Error(output.stderr || `Command failed with code ${output.code}`);
    }
    return output.stdout;
}

// ── Mock data for browser preview ──

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
    { hash: 'e7d2c8a1f5b3', message: 'Update bracket mounting holes from 4mm to 5mm', author: 'engineer', date: new Date(Date.now() - 7200000).toISOString() },
    { hash: 'b1c4f7e2a8d5', message: 'Add gasket seal profile to assembly', author: 'designer', date: new Date(Date.now() - 86400000).toISOString() },
    { hash: 'f5a2d8c1b7e3', message: 'Initial CAD design — housing + bracket', author: 'engineer', date: new Date(Date.now() - 172800000).toISOString() },
    { hash: 'c8e1a5f3d2b7', message: 'Project initialization', author: 'engineer', date: new Date(Date.now() - 259200000).toISOString() },
];

const MOCK_DIFFS: DiffResult[] = [
    {
        filePath: 'housing.stl',
        fileType: 'stl',
        summary: 'STL: 1248 triangles → 1396 triangles (+192 -44)',
        isCAD: true,
        additions: 192,
        deletions: 44,
        lines: [
            { type: 'del', content: 'Triangles: 1248' },
            { type: 'add', content: 'Triangles: 1396' },
            { type: 'del', content: 'Removed 44 triangles from mesh' },
            { type: 'add', content: 'Added 192 triangles to mesh' },
            { type: 'del', content: 'Bounding box: (0.00, 0.00, 0.00) → (50.00, 30.00, 25.00)' },
            { type: 'add', content: 'Bounding box: (0.00, 0.00, 0.00) → (52.00, 30.00, 25.00)' },
            { type: 'del', content: 'Surface area: 6240.00' },
            { type: 'add', content: 'Surface area: 6580.50' },
        ],
    },
    {
        filePath: 'bracket.dxf',
        fileType: 'dxf',
        summary: 'DXF: 34 → 38 entities (+6 -2)',
        isCAD: true,
        additions: 6,
        deletions: 2,
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
        filePath: 'assembly.obj',
        fileType: 'obj',
        summary: 'new OBJ: assembly.obj (842 vertices, 1560 faces)',
        isCAD: true,
        additions: 842,
        deletions: 0,
        lines: [
            { type: 'add', content: 'Vertices: 842' },
            { type: 'add', content: 'Faces: 1560' },
            { type: 'add', content: 'Normals: 842' },
            { type: 'add', content: 'Texture coordinates: 0' },
            { type: 'add', content: 'Group added: housing' },
            { type: 'add', content: 'Group added: bracket' },
            { type: 'add', content: 'Group added: gasket' },
        ],
    },
];

// ── API Functions ──

export async function getStatus(repoPath: string): Promise<{ branch: string; files: StatusFile[] }> {
    if (!isTauri) {
        return { branch: 'main', files: [...MOCK_STATUS_FILES] };
    }

    const output = await runGitCAD(['status'], repoPath);
    const lines = output.split('\n');
    const files: StatusFile[] = [];
    let branch = 'main';

    for (const line of lines) {
        const branchMatch = line.match(/On branch (\S+)/);
        if (branchMatch) branch = branchMatch[1];

        if (line.includes('new file:')) {
            const path = line.replace(/.*new file:\s*/, '').trim();
            if (path) files.push({ path, status: 'A' });
        } else if (line.includes('modified:')) {
            const path = line.replace(/.*modified:\s*/, '').trim();
            if (path) files.push({ path, status: 'M' });
        } else if (line.includes('deleted:')) {
            const path = line.replace(/.*deleted:\s*/, '').trim();
            if (path) files.push({ path, status: 'D' });
        }

        const untrackedMatch = line.match(/^\s{2}(\S.*)$/);
        if (untrackedMatch && !line.includes(':')) {
            files.push({ path: untrackedMatch[1].trim(), status: 'U' });
        }
    }

    return { branch, files };
}

export async function getLog(repoPath: string): Promise<CommitInfo[]> {
    if (!isTauri) return [...MOCK_COMMITS];

    try {
        const output = await runGitCAD(['log'], repoPath);
        const commits: CommitInfo[] = [];
        const blocks = output.split(/(?=commit )/);

        for (const block of blocks) {
            if (!block.trim()) continue;
            const lines = block.trim().split('\n');
            const commitLine = lines[0] || '';
            const hashMatch = commitLine.match(/commit\s+(\S+)/);
            const branchMatch = commitLine.match(/->\s*(\S+)/);
            const authorLine = lines.find(l => l.startsWith('Author:'));
            const dateLine = lines.find(l => l.startsWith('Date:'));
            const messageLine = lines.find(l => l.startsWith('    '));

            if (hashMatch) {
                commits.push({
                    hash: hashMatch[1],
                    message: messageLine?.trim() || '',
                    author: authorLine?.replace('Author:', '').trim() || '',
                    date: dateLine?.replace('Date:', '').trim() || '',
                    branch: branchMatch?.[1] || undefined,
                });
            }
        }
        return commits;
    } catch {
        return [];
    }
}

export async function getDiff(repoPath: string): Promise<DiffResult[]> {
    if (!isTauri) return [...MOCK_DIFFS];

    try {
        const output = await runGitCAD(['diff'], repoPath);
        const results: DiffResult[] = [];
        const blocks = output.split(/(?=diff --gitcad )/);

        for (const block of blocks) {
            if (!block.trim() || !block.startsWith('diff --gitcad')) continue;
            const lines = block.trim().split('\n');
            const headerMatch = lines[0].match(/diff --gitcad\s+(.+)/);
            if (!headerMatch) continue;

            const filePath = headerMatch[1];
            let fileType = 'text';
            const cadLine = lines.find(l => l.startsWith('CAD format:'));
            if (cadLine) fileType = cadLine.replace('CAD format:', '').trim();

            const summaryLine = lines.find(l =>
                !l.startsWith('diff ') && !l.startsWith('CAD format:') && !l.startsWith('─') && l.trim() !== ''
            ) || '';

            const diffLines: DiffLine[] = [];
            let additions = 0, deletions = 0;
            let pastSeparator = false;

            for (const line of lines) {
                if (line.startsWith('─')) { pastSeparator = true; continue; }
                if (!pastSeparator) continue;

                if (line.startsWith('+ ') || line.startsWith('+\t')) {
                    diffLines.push({ type: 'add', content: line.substring(2) });
                    additions++;
                } else if (line.startsWith('- ') || line.startsWith('-\t')) {
                    diffLines.push({ type: 'del', content: line.substring(2) });
                    deletions++;
                } else if (line.startsWith('  ')) {
                    diffLines.push({ type: 'ctx', content: line.substring(2) });
                }
            }

            results.push({ filePath, fileType, summary: summaryLine, lines: diffLines, isCAD: fileType !== 'text', additions, deletions });
        }
        return results;
    } catch {
        return [];
    }
}

export async function getBranches(repoPath: string): Promise<BranchInfo[]> {
    if (!isTauri) return [...MOCK_BRANCHES];

    try {
        const output = await runGitCAD(['branch'], repoPath);
        const branches: BranchInfo[] = [];
        for (const line of output.split('\n')) {
            const trimmed = line.trim();
            if (!trimmed) continue;
            const active = trimmed.startsWith('* ');
            const name = trimmed.replace(/^\*?\s*/, '');
            if (name) branches.push({ name, active });
        }
        return branches;
    } catch {
        return [];
    }
}

export async function checkoutBranch(repoPath: string, branch: string): Promise<string> {
    if (!isTauri) return `Switched to branch '${branch}'`;
    return runGitCAD(['checkout', branch], repoPath);
}

export async function createBranch(repoPath: string, name: string): Promise<string> {
    if (!isTauri) return `Created branch '${name}'`;
    return runGitCAD(['branch', name], repoPath);
}

export async function addFiles(repoPath: string, files: string[]): Promise<string> {
    if (!isTauri) return 'added files';
    return runGitCAD(['add', ...files], repoPath);
}

export async function commit(repoPath: string, message: string): Promise<string> {
    if (!isTauri) return `[main abc123] ${message}`;
    return runGitCAD(['commit', '-m', message], repoPath);
}

export async function mergeBranch(repoPath: string, branch: string): Promise<string> {
    if (!isTauri) return `Merged branch '${branch}'`;
    return runGitCAD(['merge', branch], repoPath);
}

export async function initRepo(repoPath: string): Promise<string> {
    if (!isTauri) return 'Initialized empty gitcad repository';
    return runGitCAD(['init'], repoPath);
}
