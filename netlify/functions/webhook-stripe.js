import Stripe from 'stripe';

export const handler = async (event) => {
  const stripe = new Stripe(process.env.STRIPE_SECRET_KEY);
  const sig = event.headers['stripe-signature'];
  let stripeEvent;
  try {
    stripeEvent = stripe.webhooks.constructEvent(event.body, sig, process.env.STRIPE_WEBHOOK_SECRET);
  } catch (err) {
    console.error('Stripe webhook signature error', err.message);
    return { statusCode: 400, body: `Webhook Error: ${err.message}` };
  }

  const session = stripeEvent.data.object;
  switch (stripeEvent.type) {
    case 'checkout.session.completed':
      console.log('Stripe subscription activated', session.customer_email, session.metadata?.plan);
      break;
    case 'customer.subscription.deleted':
      console.log('Stripe subscription cancelled', session.customer);
      break;
    default:
      console.log('Unhandled Stripe event', stripeEvent.type);
  }
  return { statusCode: 200, body: JSON.stringify({ received: true }) };
};
