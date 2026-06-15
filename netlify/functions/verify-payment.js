import Stripe from 'stripe';
import crypto from 'crypto';

function makeToken(email, plan) {
  const expiry = Date.now() + 32 * 24 * 60 * 60 * 1000; // 32 days
  const payload = Buffer.from(`${email}:${plan}:${expiry}`).toString('base64url');
  const sig = crypto.createHmac('sha256', process.env.TOKEN_SECRET || 'dev-secret').update(payload).digest('base64url');
  return `${payload}.${sig}`;
}

export const handler = async (event) => {
  if (event.httpMethod !== 'POST') return { statusCode: 405, body: 'Method Not Allowed' };
  let body;
  try { body = JSON.parse(event.body); } catch { return { statusCode: 400, body: 'Bad request' }; }

  const { method, sessionId, razorpayOrderId, razorpayPaymentId, razorpaySignature, plan, email } = body;

  try {
    if (method === 'stripe') {
      const stripe = new Stripe(process.env.STRIPE_SECRET_KEY);
      const session = await stripe.checkout.sessions.retrieve(sessionId);
      if (session.payment_status !== 'paid' && session.status !== 'complete') {
        return { statusCode: 402, body: 'Payment not complete' };
      }
      const resolvedPlan = session.metadata?.plan || plan;
      const resolvedEmail = session.customer_email || email;
      return { statusCode: 200, headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ token: makeToken(resolvedEmail, resolvedPlan), plan: resolvedPlan, email: resolvedEmail }) };
    }

    if (method === 'razorpay') {
      const secret = process.env.RAZORPAY_KEY_SECRET || '';
      const expectedSig = crypto.createHmac('sha256', secret).update(`${razorpayOrderId}|${razorpayPaymentId}`).digest('hex');
      if (expectedSig !== razorpaySignature) return { statusCode: 400, body: 'Signature mismatch' };
      return { statusCode: 200, headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ token: makeToken(email, plan), plan, email }) };
    }

    if (method === 'coinbase') {
      // Coinbase webhooks confirm async; issue provisional token, webhook will revoke if needed
      if (!email || !plan) return { statusCode: 400, body: 'email and plan required' };
      return { statusCode: 200, headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ token: makeToken(email, plan), plan, email }) };
    }

    return { statusCode: 400, body: 'Unknown method' };
  } catch (err) {
    console.error('verify-payment error', err);
    return { statusCode: 500, body: err.message };
  }
};
