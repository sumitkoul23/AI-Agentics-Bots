import { getStore } from '@netlify/blobs'
import {
  generateRegistrationOptions,
  verifyRegistrationResponse,
  generateAuthenticationOptions,
  verifyAuthenticationResponse,
} from '@simplewebauthn/server'
import crypto from 'crypto'

const RP_NAME = 'Bodhi AI'
const RP_ID = process.env.SITE_DOMAIN ||
  (process.env.SITE_URL ? new URL(process.env.SITE_URL).hostname : 'localhost')
const ORIGIN = process.env.SITE_URL || `https://${RP_ID}`

function makeToken(email, plan) {
  const expiry = Date.now() + 32 * 24 * 60 * 60 * 1000
  const payload = Buffer.from(`${email}:${plan}:${expiry}`).toString('base64url')
  const sig = crypto.createHmac('sha256', process.env.TOKEN_SECRET || 'dev-secret').update(payload).digest('base64url')
  return `${payload}.${sig}`
}

export const handler = async (event) => {
  if (event.httpMethod !== 'POST') return { statusCode: 405, body: 'Method Not Allowed' }

  let body
  try { body = JSON.parse(event.body) } catch { return { statusCode: 400, body: 'Bad request' } }

  const { action, email } = body
  if (!email) return { statusCode: 400, body: 'Email required' }

  const users = getStore('bodhi-users')
  const challenges = getStore('bodhi-challenges')

  try {
    if (action === 'register-start') {
      let userData = null
      try { userData = await users.get(email, { type: 'json' }) } catch {}

      const userID = userData?.userId || crypto.randomBytes(16).toString('base64url')
      const existingCredentials = userData?.credentials || []

      const options = await generateRegistrationOptions({
        rpName: RP_NAME,
        rpID: RP_ID,
        userID: Buffer.from(userID),
        userName: email,
        userDisplayName: email,
        attestationType: 'none',
        excludeCredentials: existingCredentials.map(c => ({
          id: c.credentialID,
          type: 'public-key',
          transports: c.transports,
        })),
        authenticatorSelection: {
          residentKey: 'preferred',
          userVerification: 'preferred',
        },
      })

      await challenges.setJSON(`${email}:register`, {
        challenge: options.challenge,
        userId: userID,
        expiry: Date.now() + 5 * 60 * 1000,
      })

      return {
        statusCode: 200,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(options),
      }
    }

    if (action === 'register-finish') {
      const { credential } = body
      const challengeData = await challenges.get(`${email}:register`, { type: 'json' })
      if (!challengeData || Date.now() > challengeData.expiry) {
        return { statusCode: 400, body: 'Challenge expired — please try again' }
      }

      const verification = await verifyRegistrationResponse({
        response: credential,
        expectedChallenge: challengeData.challenge,
        expectedOrigin: ORIGIN,
        expectedRPID: RP_ID,
      })
      if (!verification.verified) return { statusCode: 400, body: 'Passkey verification failed' }

      const { registrationInfo } = verification
      let userData = null
      try { userData = await users.get(email, { type: 'json' }) } catch {}
      userData = userData || { userId: challengeData.userId, email, plan: 'free', credentials: [] }

      userData.credentials.push({
        credentialID: registrationInfo.credential.id,
        credentialPublicKey: Buffer.from(registrationInfo.credential.publicKey).toString('base64url'),
        counter: registrationInfo.credential.counter,
        transports: credential.response.transports || [],
      })

      await users.setJSON(email, userData)
      await challenges.delete(`${email}:register`)

      const token = makeToken(email, userData.plan)
      return {
        statusCode: 200,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ token, plan: userData.plan, email }),
      }
    }

    if (action === 'login-start') {
      let userData = null
      try { userData = await users.get(email, { type: 'json' }) } catch {}

      const allowCredentials = (userData?.credentials || []).map(c => ({
        id: c.credentialID,
        type: 'public-key',
        transports: c.transports,
      }))

      const options = await generateAuthenticationOptions({
        rpID: RP_ID,
        userVerification: 'preferred',
        allowCredentials: allowCredentials.length > 0 ? allowCredentials : undefined,
      })

      await challenges.setJSON(`${email}:login`, {
        challenge: options.challenge,
        expiry: Date.now() + 5 * 60 * 1000,
      })

      return {
        statusCode: 200,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(options),
      }
    }

    if (action === 'login-finish') {
      const { credential } = body
      const challengeData = await challenges.get(`${email}:login`, { type: 'json' })
      if (!challengeData || Date.now() > challengeData.expiry) {
        return { statusCode: 400, body: 'Challenge expired — please try again' }
      }

      let userData = null
      try { userData = await users.get(email, { type: 'json' }) } catch {}
      if (!userData) return { statusCode: 404, body: 'No account found for this email' }

      const storedCred = userData.credentials.find(c => c.credentialID === credential.id)
      if (!storedCred) return { statusCode: 404, body: 'Passkey not found. Try a different passkey or use password.' }

      const verification = await verifyAuthenticationResponse({
        response: credential,
        expectedChallenge: challengeData.challenge,
        expectedOrigin: ORIGIN,
        expectedRPID: RP_ID,
        credential: {
          id: storedCred.credentialID,
          publicKey: Buffer.from(storedCred.credentialPublicKey, 'base64url'),
          counter: storedCred.counter,
          transports: storedCred.transports,
        },
      })
      if (!verification.verified) return { statusCode: 400, body: 'Passkey verification failed' }

      storedCred.counter = verification.authenticationInfo.newCounter
      await users.setJSON(email, userData)
      await challenges.delete(`${email}:login`)

      const token = makeToken(email, userData.plan)
      return {
        statusCode: 200,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ token, plan: userData.plan, email }),
      }
    }

    return { statusCode: 400, body: 'Unknown action' }
  } catch (err) {
    console.error('auth-passkey error', err)
    return { statusCode: 500, body: err.message }
  }
}
