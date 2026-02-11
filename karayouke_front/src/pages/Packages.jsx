import { useState, useEffect, useCallback } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth, fetchWithAuth } from '../lib/auth.jsx';
import { useCurrency } from '../lib/currency.jsx';

const API_BASE = (() => {
	const raw = import.meta.env.VITE_WS_HOST?.trim();
	if (raw && raw.length > 0) return raw.replace(/\/$/, '');
	if (typeof window !== 'undefined' && window.location) return window.location.origin.replace(/\/$/, '');
	return '';
})();

const Packages = () => {
	const { isAuthenticated, isLoading: authLoading } = useAuth();
	const navigate = useNavigate();
	const { format } = useCurrency();
	
	const [plans, setPlans] = useState([]);
	const [packages, setPackages] = useState([]);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState(null);
	const [purchasing, setPurchasing] = useState(null); // id of item being purchased

	const fetchData = useCallback(async () => {
		try {
			const [plansRes, packagesRes] = await Promise.all([
				fetch(`${API_BASE}/api/subscription-plans`),
				fetch(`${API_BASE}/api/packages`),
			]);
			if (!plansRes.ok) throw new Error('Failed to fetch subscription plans');
			if (!packagesRes.ok) throw new Error('Failed to fetch packages');
			const plansData = await plansRes.json();
			const packagesData = await packagesRes.json();
			setPlans(plansData || []);
			setPackages(packagesData || []);
		} catch (err) {
			setError(err.message);
		} finally {
			setLoading(false);
		}
	}, []);

	useEffect(() => {
		fetchData();
	}, [fetchData]);

	const handlePurchase = async (type, id) => {
		if (!isAuthenticated) {
			navigate('/login', { state: { from: '/packages' } });
			return;
		}
		setPurchasing(id);
		try {
			const body = type === 'subscription'
				? { plan_id: id }
				: { package_id: id };
			const response = await fetchWithAuth(`${API_BASE}/api/ipaymu/create-payment`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(body),
			});
			if (!response.ok) {
				const errData = await response.json();
				throw new Error(errData.error || 'Failed to create payment');
			}
			const data = await response.json();
			if (data.payment_url) {
				window.location.href = data.payment_url;
			} else {
				throw new Error('No payment URL returned');
			}
		} catch (err) {
			setError(err.message);
			setPurchasing(null);
		}
	};

	const parseDetail = (detail) => {
		if (!detail) return [];
		try {
			if (typeof detail === 'string') return JSON.parse(detail);
			if (Array.isArray(detail)) return detail;
			if (detail.length && typeof detail[0] === 'number') {
				return JSON.parse(String.fromCharCode(...detail));
			}
			return [];
		} catch {
			return [];
		}
	};

	if (authLoading) {
		return (
			<div className="packages-page">
				<div className="packages-loading"><span className="auth-spinner" /></div>
			</div>
		);
	}

	return (
		<div className="packages-page">
			<header className="packages-header">
				<div className="packages-brand">
					<Link to="/" className="packages-back"><span>←</span></Link>
					<span className="packages-logo">♪</span>
					<span className="packages-title">Karayouke</span>
				</div>
			</header>

			<main className="packages-main">
				{/* Free Plan Info */}
				<section className="packages-hero">
					<h1>Plans &amp; Pricing</h1>
					<p>Every user gets <strong>5 free credits daily</strong> and <strong>40-minute rooms</strong> on the free plan. Upgrade for more!</p>
				</section>

				{loading ? (
					<div className="packages-loading-content"><span className="auth-spinner" /></div>
				) : error ? (
					<div className="packages-error">
						<p>{error}</p>
						<button onClick={() => { setError(null); fetchData(); }} className="packages-retry-btn">Try Again</button>
					</div>
				) : (
					<>
						{/* Subscription Plans */}
						{plans.length > 0 && (
							<section className="packages-section">
								<h2>Subscription Plans</h2>
								<p className="packages-section-desc">Get more daily free credits and longer room duration.</p>
								<div className="packages-grid">
									{plans.map((plan, idx) => {
										const details = parseDetail(plan.plan_detail);
										const isPopular = idx === Math.floor(plans.length / 2);
										return (
											<div key={plan.id} className={`package-card ${isPopular ? 'popular' : ''}`}>
												{isPopular && <div className="package-badge">Popular</div>}
												<div className="package-header">
													<h3>{plan.plan_name}</h3>
													<div className="package-credits">
														<span className="credits-amount">{plan.daily_free_credits}</span>
														<span className="credits-label">daily credits</span>
													</div>
												</div>
												<div className="package-price">
													<span className="price-main">{format(plan.price)}</span>
													<span className="price-period">/ {plan.billing_period_days} days</span>
												</div>
												<ul className="package-features">
													<li><span className="feature-check">✓</span>{plan.daily_free_credits} free credits/day</li>
													<li><span className="feature-check">✓</span>{plan.room_duration_minutes}-minute rooms</li>
													{details.map((f, i) => (
														<li key={i}><span className="feature-check">✓</span>{f}</li>
													))}
												</ul>
												<button
													className="package-buy-btn"
													disabled={purchasing === plan.id}
													onClick={() => handlePurchase('subscription', plan.id)}
												>
													{purchasing === plan.id ? 'Processing...' : isAuthenticated ? 'Subscribe Now' : 'Sign in to Subscribe'}
												</button>
											</div>
										);
									})}
								</div>
							</section>
						)}

						{/* Extra Credit Packages */}
						{packages.length > 0 && (
							<section className="packages-section">
								<h2>Extra Credits</h2>
								<p className="packages-section-desc">Purchase additional credits that never expire. Used after your daily free credits run out.</p>
								<div className="packages-grid">
									{packages.map((pkg) => {
										const details = parseDetail(pkg.package_detail);
										const isBestValue = pkg.credit_amount >= 200;
										return (
											<div key={pkg.id} className={`package-card ${isBestValue ? 'best-value' : ''}`}>
												{isBestValue && <div className="package-badge best">Best Value</div>}
												<div className="package-header">
													<h3>{pkg.package_name}</h3>
													<div className="package-credits">
														<span className="credits-amount">{pkg.credit_amount}</span>
														<span className="credits-label">extra credits</span>
													</div>
												</div>
												<div className="package-price">
													<span className="price-main">{format(pkg.price)}</span>
													<span className="price-period">one-time</span>
												</div>
												{details.length > 0 && (
													<ul className="package-features">
														{details.map((f, i) => (
															<li key={i}><span className="feature-check">✓</span>{f}</li>
														))}
													</ul>
												)}
												<button
													className="package-buy-btn"
													disabled={purchasing === pkg.id}
													onClick={() => handlePurchase('extra_credit', pkg.id)}
												>
													{purchasing === pkg.id ? 'Processing...' : isAuthenticated ? 'Purchase Now' : 'Sign in to Purchase'}
												</button>
											</div>
										);
									})}
								</div>
							</section>
						)}
					</>
				)}

				<section className="packages-faq">
					<h2>Frequently Asked Questions</h2>
					<div className="faq-grid">
						<div className="faq-item">
							<h4>How do credits work?</h4>
							<p>Credits are used to create karaoke rooms. Each room costs 1 credit. Free credits reset daily, extra credits never expire.</p>
						</div>
						<div className="faq-item">
							<h4>What's the difference between free and extra credits?</h4>
							<p>Free credits are given daily based on your subscription plan (5 for free users). Extra credits are purchased and never expire. Free credits are used first.</p>
						</div>
						<div className="faq-item">
							<h4>What payment methods are accepted?</h4>
							<p>We accept bank transfers, virtual accounts, QRIS, e-wallets, and other Indonesian payment methods via iPaymu.</p>
						</div>
						<div className="faq-item">
							<h4>Can I get a refund?</h4>
							<p>Please contact support for refund requests within 7 days of purchase.</p>
						</div>
					</div>
				</section>
			</main>
		</div>
	);
};

export default Packages;
