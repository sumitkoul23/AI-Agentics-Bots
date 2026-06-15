import Stripe from 'stripe';
import Razorpay from 'razorpay';

const PLANS = {
  pro:  { usd: parseInt(process.env.STRIPE_PRICE_PRO  || '999'),  inr: 79900,  usdc: '9.00'  },
  team: { usd: parseInt(process.env.STRIPE_PRICE_TEAM || '4900'), inr: 399900, usdc: '44.00' }
};

export const handler = async (event) => {
  if (event.httpMethod !== 'POST') return { statusCode: 405, body: 'Method Not Allowed' };

  let body;
  try { body = JSON.parse(event.body); } catch { return { statusCode: 400, body: 'Bad request' }; }

  const { plan, method, email } = body;
  if (!plan || !['pro','team'].includes(plan)) return { statusCode: 400, body: 'Invalid plan' };
  if (!email) return { statusCode: 400, body: 'Email required' };
  if (!method || !['stripe','razorpay','coinbase'].includes(method)) return { statusCode: 400, body: 'Invalid method' };

  const origin = event.headers.origin || event.headers.referer?.replace(/\/$/, '') || 'https://bodhiapp.netlify.app';

  try {
    if (method === 'stripe') {
      const stripe = new Stripe(process.env.STRIPE_SECRET_KEY);
      const priceId = plan === 'pro' ? process.env.STRIPE_PRICE_PRO : process.env.STRIPE_PRICE_TEAM;
      if (!priceId) return { statusCode: 503, body: 'Stripe not configured' };

      const session = await stripe.checkout.sessions.create({
        mode: 'subscription',
        customer_email: email,
        line_items: [{ price: priceId, quantity: 1 }],
        success_url: `${origin}/payment-success?session_id={CHECKOUT_SESSION_ID}&plan=${plan}`,
        cancel_url: `${origin}/?screen=pricing`,
        metadata: { plan, email }
      });
      return { statusCode: 200, headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ url: session.url }) };
    }

    if (method === 'razorpay') {
      const rp = new Razorpay({ key_id: process.env.RAZORPAY_KEY_ID, key_secret: process.env.RAZORPAY_KEY_SECRET });
      if (!process.env.RAZORPAY_KEY_ID) return { statusCode: 503, body: 'Razorpay not configured' };

      const order = await rp.orders.create({
        amount: PLANS[plan].inr,
        currency: 'INR',
        receipt: `bodhi_${plan}_${Date.now()}`,
        notes: { plan, email }
      });
      return {
        statusCode: 200,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ orderId: order.id, amount: order.amount, currency: 'INR', keyId: process.env.RAZORPAY_KEY_ID, plan, email })
      };
    }

    if (method === 'coinbase') {
      if (!process.env.COINBASE_COMMERCE_API_KEY) return { statusCode: 503, body: 'Coinbase Commerce not configured' };
      const res = await fetch('https://api.commerce.coinbase.com/charges', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'X-CC-Api-Key': process.env.COINBASE_COMMERCE_API_KEY, 'X-CC-Version': '2018-03-22' },
        body: JSON.stringify({
          name: `Bodhi ${plan.charAt(0).toUpperCase() + plan.slice(1)} Plan`,
          description: `Monthly subscription — ${plan === 'pro' ? '35 agents, unlimited messages' : 'Pro + API access, 5 seats'}`,
          pricing_type: 'fixed_price',
          local_price: { amount: PLANS[plan].usdc, currency: 'USD' },
          metadata: { plan, email },
          redirect_url: `${origin}/payment-success?method=coinbase&plan=${plan}`,
          cancel_url: `${origin}/?screen=pricing`
        })
      });
      const charge = await res.json();
      if (!res.ok) return { statusCode: 502, body: JSON.stringify(charge) };
      return { statusCode: 200, headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ url: charge.data.hosted_url }) };
    }
  } catch (err) {
    console.error('checkout error', err);
    return { statusCode: 500, body: err.message };
  }
};
