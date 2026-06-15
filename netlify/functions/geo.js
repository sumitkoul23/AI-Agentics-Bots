export const handler = async (event) => {
  const country = event.headers['x-nf-country'] || event.headers['cf-ipcountry'] || 'US';
  const isIndia = country === 'IN';
  return {
    statusCode: 200,
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      country,
      isIndia,
      currency: isIndia ? 'INR' : 'USD',
      plans: {
        pro: {
          usd: { amount: 999, display: '$9.99/mo' },
          inr: { amount: 79900, display: '₹799/mo' },
          usdc: { amount: '9.00', display: '$9 USDC/mo' }
        },
        team: {
          usd: { amount: 4900, display: '$49/mo' },
          inr: { amount: 399900, display: '₹3,999/mo' },
          usdc: { amount: '44.00', display: '$44 USDC/mo' }
        }
      }
    })
  };
};
