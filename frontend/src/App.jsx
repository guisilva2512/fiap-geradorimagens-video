import { useEffect, useState } from 'react'
import { ArrowRight, CheckCircle2, Download, Film, LogOut, Plus, RefreshCw, Upload, XCircle } from 'lucide-react'
import { createBatch, downloadBatch, listBatches, listProcessings, login, register, session, uploadVideo } from './api'

const statusLabels = { PENDING: 'Na fila', PROCESSING: 'Processando', COMPLETED: 'Concluído', FAILED: 'Falhou' }

function App() {
  const [currentSession, setCurrentSession] = useState(session.get())
  const [selectedBatch, setSelectedBatch] = useState(null)

  function handleLogout() {
    session.clear()
    setCurrentSession(null)
    setSelectedBatch(null)
  }

  if (!currentSession) return <AuthScreen onAuthenticated={setCurrentSession} />
  return <Workspace user={currentSession.user} token={currentSession.token} selectedBatch={selectedBatch} onSelectBatch={setSelectedBatch} onLogout={handleLogout} />
}

function AuthScreen({ onAuthenticated }) {
  const [mode, setMode] = useState('login')
  const [form, setForm] = useState({ name: '', email: '', password: '' })
  const [feedback, setFeedback] = useState({ type: '', message: '' })
  const [loading, setLoading] = useState(false)

  async function submit(event) {
    event.preventDefault()
    setFeedback({ type: '', message: '' })
    setLoading(true)
    try {
      if (mode === 'register') {
        await register(form.name, form.email, form.password)
        setMode('login')
        setFeedback({ type: 'success', message: 'Conta criada. Entre para começar.' })
      } else {
        const data = await login(form.email, form.password)
        const nextSession = { token: data.token, user: data.user }
        session.save(nextSession)
        onAuthenticated(nextSession)
      }
    } catch (error) {
      setFeedback({ type: 'error', message: error.message })
    } finally {
      setLoading(false)
    }
  }

  return (
    <main className="auth-shell">
      <section className="auth-art" aria-label="Frameflow">
        <div className="brand-mark"><Film size={20} /> FRAMEFLOW</div>
        <div className="art-copy"><span className="eyebrow">VIDEO TO VISUAL</span><h1>Transforme movimento em momentos.</h1><p>Extraia os melhores frames dos seus vídeos em lotes organizados e prontos para baixar.</p></div>
        <div className="art-meta"><span>01</span><span className="art-line" /><span>PROCESSAMENTO ASSÍNCRONO</span></div>
      </section>
      <section className="auth-panel">
        <div className="auth-form-wrap">
          <span className="eyebrow">{mode === 'login' ? 'BEM-VINDO DE VOLTA' : 'COMECE AGORA'}</span>
          <h2>{mode === 'login' ? 'Acesse seu espaço.' : 'Crie seu espaço.'}</h2>
          <p className="muted">{mode === 'login' ? 'Entre para acompanhar seus processamentos.' : 'Organize seus vídeos e transforme cada frame.'}</p>
          {feedback.message && <div className={`feedback ${feedback.type}`}>{feedback.message}</div>}
          <form onSubmit={submit}>
            {mode === 'register' && <label>Nome<input required value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} placeholder="Seu nome" /></label>}
            <label>E-mail<input required type="email" value={form.email} onChange={e => setForm({ ...form, email: e.target.value })} placeholder="voce@exemplo.com" /></label>
            <label>Senha<input required minLength="6" type="password" value={form.password} onChange={e => setForm({ ...form, password: e.target.value })} placeholder="Mínimo de 6 caracteres" /></label>
            <button className="primary-button" disabled={loading}>{loading ? 'Aguarde...' : mode === 'login' ? <>Entrar <ArrowRight size={17} /></> : 'Criar conta'}</button>
          </form>
          <p className="switch-auth">{mode === 'login' ? 'Ainda não tem uma conta?' : 'Já tem uma conta?'} <button onClick={() => { setMode(mode === 'login' ? 'register' : 'login'); setFeedback({ type: '', message: '' }) }}>{mode === 'login' ? 'Cadastre-se' : 'Entrar'}</button></p>
        </div>
      </section>
    </main>
  )
}

function Workspace({ user, token, selectedBatch, onSelectBatch, onLogout }) {
  const [batches, setBatches] = useState([])
  const [loading, setLoading] = useState(true)
  const [creating, setCreating] = useState(false)
  const [error, setError] = useState('')

  async function refresh() {
    setLoading(true)
    try { setBatches(await listBatches(token)); setError('') } catch (e) { setError(e.message) } finally { setLoading(false) }
  }
  useEffect(() => { refresh() }, [])

  async function newBatch() {
    setCreating(true)
    try { onSelectBatch(await createBatch(token, user.id)); await refresh() } catch (e) { setError(e.message) } finally { setCreating(false) }
  }

  return <main className="app-shell">
    <header className="topbar"><div className="brand-mark dark"><Film size={19} /> FRAMEFLOW</div><div className="user-menu"><span>{user.name}</span><button className="icon-button" title="Sair" onClick={onLogout}><LogOut size={18} /></button></div></header>
    <div className="content-grid">
      <aside className="sidebar"><div className="sidebar-heading"><div><span className="eyebrow">SEU ARQUIVO</span><h1>Lotes</h1></div><button className="new-button" title="Novo lote" onClick={newBatch} disabled={creating}><Plus size={18} /></button></div><p className="muted small">Cada lote reúne vídeos e imagens de um mesmo trabalho.</p>{error && <div className="feedback error">{error}</div>}<div className="batch-list">{loading ? <div className="empty-state">Carregando lotes...</div> : batches.length === 0 ? <div className="empty-state">Nenhum lote ainda.<br />Crie o primeiro para começar.</div> : batches.map(batch => <button className={`batch-item ${selectedBatch?.id === batch.id ? 'active' : ''}`} key={batch.id} onClick={() => onSelectBatch(batch)}><span className="batch-icon"><Film size={16} /></span><span><strong>Lote {batch.id.slice(0, 8)}</strong><small>{formatDate(batch.created_at)}</small></span><ArrowRight size={15} /></button>)}</div></aside>
      <section className="detail"><div className="detail-header"><div><span className="eyebrow">{selectedBatch ? 'DETALHE DO LOTE' : 'PAINEL DE CONTROLE'}</span><h2>{selectedBatch ? `Lote ${selectedBatch.id.slice(0, 8)}` : 'Seu próximo frame começa aqui.'}</h2></div>{selectedBatch && <button className="secondary-button" onClick={refresh}><RefreshCw size={16} /> Atualizar lotes</button>}</div>{selectedBatch ? <BatchDetail batch={selectedBatch} token={token} /> : <EmptyDetail onCreate={newBatch} creating={creating} />}</section>
    </div>
  </main>
}

function BatchDetail({ batch, token }) {
  const [processings, setProcessings] = useState([])
  const [loading, setLoading] = useState(true)
  const [uploading, setUploading] = useState(false)
  const [message, setMessage] = useState({ type: '', text: '' })
  const [downloading, setDownloading] = useState(false)

  async function refresh() { try { setProcessings(await listProcessings(token, batch.id)) } catch (e) { setMessage({ type: 'error', text: e.message }) } finally { setLoading(false) } }
  useEffect(() => { setLoading(true); refresh() }, [batch.id])
  useEffect(() => { if (!processings.some(item => ['PENDING', 'PROCESSING'].includes(item.status))) return undefined; const timer = setInterval(refresh, 5000); return () => clearInterval(timer) }, [processings, batch.id])

  async function handleUpload(event) { const file = event.target.files[0]; if (!file) return; setUploading(true); setMessage({ type: '', text: '' }); try { await uploadVideo(token, batch.id, file); setMessage({ type: 'success', text: 'Vídeo enviado e colocado na fila.' }); await refresh() } catch (e) { setMessage({ type: 'error', text: e.message }) } finally { setUploading(false); event.target.value = '' } }
  async function handleDownload() { setDownloading(true); try { await downloadBatch(token, batch.id) } catch (e) { setMessage({ type: 'error', text: e.message }) } finally { setDownloading(false) } }
  const hasCompleted = processings.some(item => item.status === 'COMPLETED')

  return <div className="batch-detail"><div className="batch-meta"><span>ID {batch.id}</span><span>Criado em {formatDate(batch.created_at)}</span></div><div className="upload-zone"><div className="upload-symbol"><Upload size={22} /></div><div><strong>{uploading ? 'Enviando vídeo...' : 'Adicione um vídeo ao lote'}</strong><p>MP4, AVI, MOV, MKV, WMV, FLV ou WEBM</p></div><label className="secondary-button upload-button"><Upload size={16} /> Selecionar vídeo<input type="file" accept=".mp4,.avi,.mov,.mkv,.wmv,.flv,.webm,video/*" onChange={handleUpload} disabled={uploading} /></label></div>{message.text && <div className={`feedback ${message.type}`}>{message.text}</div>}<div className="processing-heading"><div><span className="eyebrow">ATIVIDADE</span><h3>Processamentos <span>{processings.length}</span></h3></div>{hasCompleted && <button className="secondary-button" onClick={handleDownload} disabled={downloading}><Download size={16} /> {downloading ? 'Preparando...' : 'Baixar imagens'}</button>}</div>{loading ? <div className="empty-state">Consultando status...</div> : processings.length === 0 ? <div className="empty-state bordered">Os vídeos enviados aparecerão aqui.</div> : <div className="processing-list">{processings.map(item => <ProcessingRow key={item.id} item={item} />)}</div>}</div>
}

function ProcessingRow({ item }) { const status = item.status || 'PENDING'; return <div className="processing-row"><div className={`status-icon ${status.toLowerCase()}`}>{status === 'COMPLETED' ? <CheckCircle2 size={17} /> : status === 'FAILED' ? <XCircle size={17} /> : <RefreshCw size={17} />}</div><div className="processing-name"><strong>{item.name}</strong><small>{formatDate(item.created_at)}</small></div><span className={`status-pill ${status.toLowerCase()}`}>{statusLabels[status] || status}</span>{status === 'FAILED' && <small className="failure">{item.error_message || 'Falha no processamento'}</small>}</div> }
function EmptyDetail({ onCreate, creating }) { return <div className="welcome"><div className="welcome-orbit"><Film size={34} /></div><h3>Um espaço para cada história.</h3><p>Crie um lote, envie seu vídeo e acompanhe a extração de imagens em tempo real.</p><button className="primary-button compact" onClick={onCreate} disabled={creating}><Plus size={17} /> {creating ? 'Criando...' : 'Criar primeiro lote'}</button></div> }
function formatDate(value) { if (!value) return 'Data não disponível'; const date = new Date(value.replace(' ', 'T')); return Number.isNaN(date.getTime()) ? value : date.toLocaleDateString('pt-BR', { day: '2-digit', month: 'short', year: 'numeric' }) }

export default App
