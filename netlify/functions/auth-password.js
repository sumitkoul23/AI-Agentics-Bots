import { getStore } from '@netlify/blobs'
import crypto from 'crypto'

function makeToken(email, plan) {
  const expiry = Date.now() + 32 * 24 * 60 * 60 * 1000
  const payload = Buffer.from(`${email}:${plan}:${expiry}`).toString('base64url')
  const sig = crypto.createHmac('sha256', process.env.TOKEN_SECRET || 'dev-secret').update(payload).digest('base64url')
  return `${payload}.${sig}`
}

function hashPassword(password, salt) {
  return new Promise((resolve, reject) => {
    crypto.pbkdf2(password, salt, 120000, 64, 'sha512', (err, key) => {
      if (err) reject(err); else resolve(key.toString('hex'))
    })
  })
}

export const handler = async (event) => {
  if (event.httpMethod !== 'POST') return { statusCode: 405, body: 'Method Not Allowed' }

  let body
  try { body = JSON.parse(event.body) } catch { return { statusCode: 400, body: 'Bad request' } }

  const { action, email, password } = body
  if (!email || !password) return { statusCode: 400, body: 'Email and password required' }
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) return { statusCode: 400, body: 'Invalid email address' }
  if (password.length < 8) return { statusCode: 400, body: 'Password must be at least 8 characters' }

  const users = getStore('bodhi-users')

  try {
    if (action === 'register') {
      let userData = null
      try { userData = await users.get(email, { type: 'json' }) } catch {}
      if (userData?.passwordHash) {
        return { statusCode: 409, body: 'Account already exists — please sign in' }
      }

      const salt = crypto.randomBytes(16).toString('hex')
      const hash = await hashPassword(password, salt)

      userData = userData || {
        userId: crypto.randomBytes(16).toString('base64url'),
        email,
        plan: 'free',
        credentials: [],
      }
      userData.passwordHash = `${salt}:${hash}`
      await users.setJSON(email, userData)

      const token = makeToken(email, userData.plan)
      return {
        statusCode: 200,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ token, plan: userData.plan, email }),
      }
    }

    if (action === 'login') {
      let userData = null
      try { userData = await users.get(email, { type: 'json' }) } catch {}
      if (!userData?.passwordHash) {
        return { statusCode: 401, body: 'Invalid email or password' }
      }

      const [salt, storedHash] = userData.passwordHash.split(':')
      const hash = await hashPassword(password, salt)

      // Constant-time comparison
      const hashBuf = Buffer.from(hash, 'hex')
      const storedBuf = Buffer.from(storedHash, 'hex')
      if (hashBuf.length !== storedBuf.length || !crypto.timingSafeEqual(hashBuf, storedBuf)) {
        return { statusCode: 401, body: 'Invalid email or password' }
      }

      const token = makeToken(email, userData.plan)
      return {
        statusCode: 200,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ token, plan: userData.plan, email }),
      }
    }

    return { statusCode: 400, body: 'Unknown action' }
  } catch (err) {
    console.error('auth-password error', err)
    return { statusCode: 500, body: err.message }
  }
}
