const AUTH_API_URL = import.meta.env.VITE_AUTH_API_URL || 'http://localhost:8081'
const UPLOAD_API_URL = import.meta.env.VITE_UPLOAD_API_URL || 'http://localhost:8082'

export const session = {
  get() {
    try {
      return JSON.parse(localStorage.getItem('frameflow_session'))
    } catch {
      return null
    }
  },
  save(value) {
    localStorage.setItem('frameflow_session', JSON.stringify(value))
  },
  clear() {
    localStorage.removeItem('frameflow_session')
  },
}

async function request(url, options = {}) {
  const response = await fetch(url, options)
  const contentType = response.headers.get('content-type') || ''
  const body = contentType.includes('application/json') ? await response.json() : await response.text()
  if (!response.ok) {
    throw new Error(body?.error || 'Não foi possível concluir a operação.')
  }
  return body
}

function authHeaders(token, headers = {}) {
  return { ...headers, Authorization: `Bearer ${token}` }
}

export async function login(email, password) {
  const response = await request(`${AUTH_API_URL}/v1/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  })
  return response.data
}

export async function register(name, email, password) {
  const response = await request(`${AUTH_API_URL}/v1/users`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, email, password }),
  })
  return response.data
}

export async function listBatches(token) {
  const response = await request(`${UPLOAD_API_URL}/v1/uploads`, { headers: authHeaders(token) })
  return response.data || []
}

export async function createBatch(token, userId) {
  const response = await request(`${UPLOAD_API_URL}/v1/uploads`, {
    method: 'POST',
    headers: authHeaders(token, { 'Content-Type': 'application/json' }),
    body: JSON.stringify({ user_id: userId }),
  })
  return response.data
}

export async function listProcessings(token, batchId) {
  const response = await request(`${UPLOAD_API_URL}/v1/uploads/${batchId}/processings`, {
    headers: authHeaders(token),
  })
  return response.data || []
}

export async function uploadVideo(token, batchId, file) {
  const formData = new FormData()
  formData.append('file', file)
  const response = await request(`${UPLOAD_API_URL}/v1/uploads/${batchId}/processings`, {
    method: 'POST',
    headers: authHeaders(token),
    body: formData,
  })
  return response.data
}

export async function downloadBatch(token, batchId) {
  const response = await fetch(`${UPLOAD_API_URL}/v1/uploads/${batchId}/download`, {
    headers: authHeaders(token),
  })
  if (!response.ok) throw new Error('O download ainda não está disponível.')
  const blob = await response.blob()
  const url = URL.createObjectURL(blob)
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = `batch-${batchId}-images.zip`
  anchor.click()
  URL.revokeObjectURL(url)
}
