import { useEffect, useState } from 'react';
import { Activity, ArrowUpRight, Blocks, Bot, CheckCircle2, CircleHelp, GitBranch, Plus, Workflow } from 'lucide-react';

type HealthStatus = 'checking' | 'online' | 'offline';

export default function App() {
  const [health, setHealth] = useState<HealthStatus>('checking');

  useEffect(() => {
    const controller = new AbortController();

    fetch('/api/health', { signal: controller.signal })
      .then((response) => {
        if (!response.ok) throw new Error('API health check failed');
        return response.json();
      })
      .then(() => setHealth('online'))
      .catch((error: unknown) => {
        if (error instanceof Error && error.name === 'AbortError') return;
        setHealth('offline');
      });

    return () => controller.abort();
  }, []);

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <a className="brand" href="#home" aria-label="FlowPilot home">
          <span className="brand-mark"><Workflow size={21} /></span>
          <span>flowpilot<span className="brand-period">.</span></span>
        </a>
        <div className="workspace-label">WORKSPACE</div>
        <nav aria-label="Main navigation">
          <a className="nav-item active" href="#workflows"><GitBranch size={17} /> Workflows <span className="nav-count">0</span></a>
          <a className="nav-item" href="#executions"><Activity size={17} /> Executions</a>
          <a className="nav-item" href="#integrations"><Blocks size={17} /> Integrations</a>
          <a className="nav-item" href="#ai-steps"><Bot size={17} /> AI steps</a>
        </nav>
        <div className="sidebar-bottom">
          <div className="api-status"><span className={`status-dot ${health}`} /> <span>API {health === 'checking' ? 'checking' : health}</span></div>
          <a className="nav-item" href="#help"><CircleHelp size={17} /> Documentation</a>
          <div className="profile"><div className="avatar">FP</div><div><strong>FlowPilot workspace</strong><span>Developer plan</span></div></div>
        </div>
      </aside>

      <main className="main-content" id="home">
        <header className="topbar"><div className="breadcrumb">Workspaces <span>/</span> <strong>Workflows</strong></div><div className="topbar-right"><span className="environment"><span /> Development</span><button className="icon-button" aria-label="Help"><CircleHelp size={18} /></button></div></header>
        <section className="page-content" id="workflows">
          <div className="eyebrow"><span className="eyebrow-line" /> AUTOMATION WORKSPACE</div>
          <div className="page-heading"><div><h1>Workflows</h1><p>Build, run, and monitor your automated workflows.</p></div><button className="primary-button" disabled title="Workflow creation is coming in the next milestone"><Plus size={17} /> New workflow <ArrowUpRight size={15} /></button></div>

          <div className="stats-grid">
            <article className="stat-card"><div className="stat-top"><span>Total workflows</span><GitBranch size={17} /></div><div className="stat-value">0</div><div className="stat-foot">Your automations, in one place</div></article>
            <article className="stat-card"><div className="stat-top"><span>Executions today</span><Activity size={17} /></div><div className="stat-value">0</div><div className="stat-foot">Runs started in this workspace</div></article>
            <article className="stat-card"><div className="stat-top"><span>Success rates</span><CheckCircle2 size={17} /></div><div className="stat-value">—</div><div className="stat-foot">Available after your first run</div></article>
          </div>

          <section className="workflow-panel">
            <div className="panel-heading"><div><h2>Your workflows</h2><p>Reusable automations created in this workspace</p></div><span className="count-pill">0 workflows</span></div>
            <div className="empty-state"><div className="empty-art"><div className="art-node node-trigger"><Activity size={18} /></div><div className="art-connector connector-one" /><div className="art-node node-ai"><Bot size={20} /></div><div className="art-connector connector-two" /><div className="art-node node-action"><CheckCircle2 size={19} /></div><div className="art-spark spark-one">✳</div><div className="art-spark spark-two">✳</div></div><h3>Your next great workflow starts here</h3><p>Connect a trigger, add steps, and let FlowPilot handle the busywork. Your workflows will appear here once creation is available.</p><button className="secondary-button" disabled> <Plus size={16} /> Create your first workflow</button><span className="coming-soon">Workflow editor · Coming in the next milestone</span></div>
          </section>
          <footer className="page-footer"><span>FlowPilot <span className="footer-dot">●</span> Early development</span><span>Built for reliable automation</span></footer>
        </section>
      </main>
    </div>
  );
}
