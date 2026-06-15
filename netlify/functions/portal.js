import Stripe from 'stripe';

export const handler = async (event) => {
  if (event.httpMethod !== 'POST') return { statusCode: 405, body: 'Method Not Allowed' };
  let body;
  try { body = JSON.parse(event.body); } catch { return { statusCode: 400, body: 'Bad request' }; }

  const { email } = body;
  if (!email) return { statusCode: 400, body: 'Email required' };
  if (!process.env.STRIPE_SECRET_KEY) return { statusCode: 503, body: 'Not configured' };

  const stripe = new Stripe(process.env.STRIPE_SECRET_KEY);
  const customers = await stripe.customers.list({ email, limit: 1 });
  if (!customers.data.length) return { statusCode: 404, body: 'No subscription found for this email' };

  const origin = event.headers.origin || 'https://bodhiapp.netlify.app';
  const session = await stripe.billingPortal.sessions.create({
    customer: customers.data[0].id,
    return_url: `${origin}/?screen=settings`
  });
  return { statusCode: 200, headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ url: session.url }) };
};
