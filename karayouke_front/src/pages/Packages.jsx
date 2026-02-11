import { useState, useEffect, useCallback } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../lib/auth.jsx';
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
	const { currency, formatFromUSD, info: currencyInfo } = useCurrency();
	
	const [packages, setPackages] = useState([]);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState(null);

	const fetchPackages = useCallback(async () => {
		try {
			const response = await fetch(`${API_BASE}/api/packages`);
			if (!response.ok) throw new Error('Failed to fetch packages');
			const data = await response.json();
			setPackages(data || []);
		} catch (err) {
			setError(err.message);
		} finally {
			setLoading(false);
		}
	}, []);

	useEffect(() => {
		fetchPackages();
	}, [fetchPackages]);

	const handlePurchase = (pkg) => {
		if (!isAuthenticated) {
			navigate('/login', { state: { from: '/packages', packageId: pkg.id } });
			return;
		}
		navigate(`/checkout/${pkg.id}`);
	};

	const parsePackageDetail = (detail) => {
		if (!detail) return [];
		try {
			if (typeof detail === 'string') {
				return JSON.parse(detail);
			}
			if (Array.isArray(detail)) {
				return detail;
			}
			// Handle byte array from Go backend
			if (detail.length && typeof detail[0] === 'number') {
				const decoded = String.fromCharCode(...detail);
				return JSON.parse(decoded);
			}
			return [];
		} catch {
			return [];
		}
	};

	const getPriceInUSD = (price) => {
		// Price is stored in cents or smallest unit, convert to dollars
		const numPrice = typeof price === 'string' ? parseFloat(price) : price;
		return numPrice / 100;
	};

	if (authLoading) {
		return (
			<div className="packages-page">
				<div className="packages-loading">
					<span className="auth-spinner" />
				</div>
			</div>
		);
	}

	return (
		<div className="packages-page">
			<header className="packages-header">
				<div className="packages-brand">
					<Link to="/" className="packages-back">
						<span>←</span>
					</Link>
					<span className="packages-logo">♪</span>
					<span className="packages-title">Karayouke</span>
				</div>
				<div className="packages-currency">
					<span className="packages-currency-label">Currency:</span>
					<span className="packages-currency-value">{currencyInfo.code}</span>
				</div>
			</header>

			<main className="packages-main">
				<section className="packages-hero">
					<h1>Credit Packages</h1>
					<p>Purchase credits to create karaoke rooms and enjoy unlimited sessions.</p>
				</section>

				{loading ? (
					<div className="packages-loading-content">
						<span className="auth-spinner" />
					</div>
				) : error ? (
					<div className="packages-error">
						<p>{error}</p>
						<button onClick={fetchPackages} className="packages-retry-btn">
							Try Again
						</button>
					</div>
				) : packages.length === 0 ? (
					<div className="packages-empty">
						<p>No packages available at the moment.</p>
					</div>
				) : (
					<div className="packages-grid">
						{packages.map((pkg) => {
							const details = parsePackageDetail(pkg.package_detail);
							const priceUSD = getPriceInUSD(pkg.price);
							const isPopular = pkg.credit_amount >= 50 && pkg.credit_amount < 200;
							const isBestValue = pkg.credit_amount >= 200;

							return (
								<div
									key={pkg.id}
									className={`package-card ${isPopular ? 'popular' : ''} ${isBestValue ? 'best-value' : ''}`}
								>
									{isPopular && <div className="package-badge">Popular</div>}
									{isBestValue && <div className="package-badge best">Best Value</div>}
									
									<div className="package-header">
										<h3>{pkg.package_name}</h3>
										<div className="package-credits">
											<span className="credits-amount">{pkg.credit_amount}</span>
											<span className="credits-label">Credits</span>
										</div>
									</div>

									<div className="package-price">
										<span className="price-main">{formatFromUSD(priceUSD)}</span>
										{currency !== 'USD' && (
											<span className="price-usd">≈ ${priceUSD.toFixed(2)} USD</span>
										)}
									</div>

									{details.length > 0 && (
										<ul className="package-features">
											{details.map((feature, idx) => (
												<li key={idx}>
													<span className="feature-check">✓</span>
													{feature}
												</li>
											))}
										</ul>
									)}

									<button
										className="package-buy-btn"
										onClick={() => handlePurchase(pkg)}
									>
										{isAuthenticated ? 'Purchase Now' : 'Sign in to Purchase'}
									</button>
								</div>
							);
						})}
					</div>
				)}

				<section className="packages-faq">
					<h2>Frequently Asked Questions</h2>
					<div className="faq-grid">
						<div className="faq-item">
							<h4>How do credits work?</h4>
							<p>Credits are used to create karaoke rooms. Each room creation costs 1 credit by default.</p>
						</div>
						<div className="faq-item">
							<h4>Do credits expire?</h4>
							<p>No, your credits never expire. Use them whenever you want.</p>
						</div>
						<div className="faq-item">
							<h4>What payment methods are accepted?</h4>
							<p>We accept major credit cards, bank transfers, and local payment methods.</p>
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
