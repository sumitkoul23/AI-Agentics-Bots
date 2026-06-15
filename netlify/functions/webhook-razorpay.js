import crypto from 'crypto';

export const handler = async (event) => {
  const sig = event.headers['x-razorpay-signature'];
  const secret = process.env.RAZORPAY_WEBHOOK_SECRET || '';
  const expected = crypto.createHmac('sha256', secret).update(event.body).digest('hex');
  if (expected !== sig) {
    console.error('Razorpay webhook signature mismatch');
    return { statusCode: 400, body: 'Signature mismatch' };
  }

  let payload;
  try { payload = JSON.parse(event.body); } catch { return { statusCode: 400, body: 'Bad body' }; }

  if (payload.event === 'payment.captured') {
    const notes = payload.payload?.payment?.entity?.notes || {};
    console.log('Razorpay payment captured', notes.email, notes.plan);
  }
  return { statusCode: 200, body: JSON.stringify({ received: true }) };
};
