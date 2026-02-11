import { useState, useEffect, useCallback } from 'react';
import { Link, useParams, useNavigate } from 'react-router-dom';
import { loadStripe } from '@stripe/stripe-js';
import { Elements, PaymentElement, useStripe, useElements } from '@stripe/react-stripe-js';
import { useAuth, fetchWithAuth, getAuthToken } from '../lib/auth.jsx';
import { useCurrency } from '../lib/currency.jsx';

const API_BASE = (() => {
	const raw = import.meta.env.VITE_WS_HOST?.trim();
	if (raw && raw.length > 0) return raw.replace(/\/$/, '');
	if (typeof window !== 'undefined' && window.location) return window.location.origin.replace(/\/$/, '');
	return '';
})();

// Stripe promise - will be initialized with publishable key
let stripePromise = null;

const getStripePromise = async () => {
	if (!stripePromise) {
		try {
			const response = await fetch(`${API_BASE}/api/stripe/config`);
			const { publishableKey } = await response.json();
			stripePromise = loadStripe(publishableKey);
		} catch (err) {
			console.error('Failed to load Stripe config:', err);
		}
	}
	return stripePromise;
};

// Stripe Payment Form Component
const StripePaymentForm = ({ transactionId, packageName, credits, onSuccess, onError }) => {
	const stripe = useStripe();
	const elements = useElements();
	const [processing, setProcessing] = useState(false);
	const [errorMessage, setErrorMessage] = useState(null);

	const handleSubmit = async (e) => {
		e.preventDefault();

		if (!stripe || !elements) {
			return;
		}

		setProcessing(true);
		setErrorMessage(null);

		try {
			const { error, paymentIntent } = await stripe.confirmPayment({
				elements,
				confirmParams: {
					return_url: `${window.location.origin}/payment/status/${transactionId}`,
				},
				redirect: 'if_required',
			});

			if (error) {
				setErrorMessage(error.message);
				onError(error.message);
			} else if (paymentIntent && paymentIntent.status === 'succeeded') {
				// Payment succeeded - confirm with backend
				try {
					await fetchWithAuth(`${API_BASE}/api/stripe/confirm-payment`, {
						method: 'POST',
						headers: { 'Content-Type': 'application/json' },
						body: JSON.stringify({
							payment_intent_id: paymentIntent.id,
							transaction_id: transactionId,
						}),
					});
				} catch (confirmErr) {
					console.error('Confirm error:', confirmErr);
				}
				onSuccess(transactionId, packageName, credits);
			} else if (paymentIntent && paymentIntent.status === 'requires_action') {
				// 3D Secure or other action required - Stripe will handle redirect
				setErrorMessage('Additional authentication required. Please complete the verification.');
			}
		} catch (err) {
			setErrorMessage(err.message || 'An unexpected error occurred');
			onError(err.message);
		} finally {
			setProcessing(false);
		}
	};

	return (
		<form onSubmit={handleSubmit} className="stripe-payment-form">
			<PaymentElement
				options={{
					layout: 'tabs',
				}}
			/>
			{errorMessage && <div className="checkout-error">{errorMessage}</div>}
			<button
				type="submit"
				className="checkout-pay-btn"
				disabled={!stripe || processing}
			>
				{processing ? 'Processing...' : 'Complete Payment'}
			</button>
		</form>
	);
};

const Checkout = () => {
	const { packageId } = useParams();
	const navigate = useNavigate();
	const { user, isAuthenticated, isLoading: authLoading } = useAuth();
	const { currency, formatFromUSD } = useCurrency();

	const [pkg, setPkg] = useState(null);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState(null);
	const [step, setStep] = useState('review'); // review, payment, processing
	const [agreeTerms, setAgreeTerms] = useState(false);

	// Stripe state
	const [stripePromiseState, setStripePromiseState] = useState(null);
	const [clientSecret, setClientSecret] = useState(null);
	const [transactionId, setTransactionId] = useState(null);

	const fetchPackage = useCallback(async () => {
		try {
			const response = await fetch(`${API_BASE}/api/packages`);
			if (!response.ok) throw new Error('Failed to fetch package');
			const packages = await response.json();
			const found = packages.find((p) => p.id === packageId);
			if (!found) throw new Error('Package not found');
			setPkg(found);
		} catch (err) {
			setError(err.message);
		} finally {
			setLoading(false);
		}
	}, [packageId]);

	useEffect(() => {
		// Check for auth token - don't rely solely on isAuthenticated state
		// This handles cases where auth state is still loading
		const hasToken = !!getAuthToken();
		
		if (!authLoading && !isAuthenticated && !hasToken) {
			navigate('/login', { state: { from: `/checkout/${packageId}` } });
			return;
		}

		if (isAuthenticated || hasToken) {
			fetchPackage();
			// Load Stripe
			getStripePromise().then(setStripePromiseState);
		}
	}, [packageId, isAuthenticated, authLoading, navigate, fetchPackage]);

	const getPriceInUSD = (price) => {
		const numPrice = typeof price === 'string' ? parseFloat(price) : price;
		return numPrice / 100;
	};

	const getCurrencyCode = () => {
		switch (currency) {
			case 'JPY':
				return 'jpy';
			case 'IDR':
				return 'idr';
			default:
				return 'usd';
		}
	};

	const createPaymentIntent = async () => {
		setError(null);
		setStep('processing');

		try {
			const response = await fetchWithAuth(`${API_BASE}/api/stripe/create-payment-intent`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					package_id: packageId,
					// Currency is now automatically detected from IP on backend
				}),
			});

			if (!response.ok) {
				const errData = await response.json();
				throw new Error(errData.error || 'Failed to create payment');
			}

			const data = await response.json();
			
			// Handle free packages - no payment needed
			if (data.free) {
				handlePaymentSuccess(data.transactionId, pkg?.package_name, pkg?.credit_amount);
				return;
			}
			
			setClientSecret(data.clientSecret);
			setTransactionId(data.transactionId);
			setStep('payment');
		} catch (err) {
			setError(err.message);
			setStep('review');
		}
	};

	const handleProceedToPayment = () => {
		if (!agreeTerms) {
			setError('Please agree to the terms and conditions');
			return;
		}
		createPaymentIntent();
	};

	const handlePaymentSuccess = (txId, pkgName, credits) => {
		navigate(`/payment/status/${txId}`, {
			state: {
				success: true,
				packageName: pkgName,
				credits: credits,
			},
		});
	};

	const handlePaymentError = (message) => {
		setError(message);
	};

	if (authLoading || loading) {
		return (
			<div className="checkout-page">
				<div className="checkout-loading">
					<span className="auth-spinner" />
				</div>
			</div>
		);
	}

	if (error && !pkg) {
		return (
			<div className="checkout-page">
				<div className="checkout-error-page">
					<h2>Error</h2>
					<p>{error}</p>
					<Link to="/packages" className="checkout-back-btn">
						Back to Packages
					</Link>
				</div>
			</div>
		);
	}

	const priceUSD = pkg ? getPriceInUSD(pkg.price) : 0;

	const stripeAppearance = {
		theme: 'night',
		variables: {
			colorPrimary: '#6366f1',
			colorBackground: '#0f172a',
			colorText: '#f8fafc',
			colorDanger: '#ef4444',
			fontFamily: 'system-ui, sans-serif',
			spacingUnit: '4px',
			borderRadius: '8px',
		},
		rules: {
			'.Input': {
				backgroundColor: 'rgba(2, 6, 23, 0.7)',
				border: '1px solid rgba(148, 163, 184, 0.2)',
			},
			'.Input:focus': {
				borderColor: '#6366f1',
			},
			'.Label': {
				color: 'rgba(226, 232, 240, 0.8)',
			},
			'.Tab': {
				backgroundColor: 'rgba(2, 6, 23, 0.5)',
				border: '1px solid rgba(148, 163, 184, 0.2)',
			},
			'.Tab--selected': {
				backgroundColor: 'rgba(99, 102, 241, 0.1)',
				borderColor: '#6366f1',
			},
		},
	};

	return (
		<div className="checkout-page">
			<header className="checkout-header">
				<Link to="/packages" className="checkout-back">
					<span>←</span> Back to Packages
				</Link>
				<div className="checkout-steps">
					<span className={`checkout-step ${step === 'review' ? 'active' : 'completed'}`}>
						1. Review
					</span>
					<span className={`checkout-step ${step === 'payment' ? 'active' : step === 'processing' ? 'completed' : ''}`}>
						2. Payment
					</span>
					<span className={`checkout-step ${step === 'processing' && !clientSecret ? 'active' : ''}`}>
						3. Complete
					</span>
				</div>
			</header>

			<main className="checkout-main">
				<div className="checkout-content">
					{/* Order Summary */}
					<div className="checkout-summary">
						<h2>Order Summary</h2>
						<div className="checkout-package-card">
							<div className="checkout-package-info">
								<h3>{pkg?.package_name}</h3>
								<p>{pkg?.credit_amount} Credits</p>
							</div>
							<div className="checkout-package-price">
								<span className="price-main">{formatFromUSD(priceUSD)}</span>
								{currency !== 'USD' && (
									<span className="price-usd">≈ ${priceUSD.toFixed(2)} USD</span>
								)}
							</div>
						</div>
						<div className="checkout-total">
							<span>Total</span>
							<span className="checkout-total-amount">{formatFromUSD(priceUSD)}</span>
						</div>
					</div>

					{/* Review Step */}
					{step === 'review' && (
						<div className="checkout-section">
							<h2>Confirm Your Purchase</h2>
							<div className="checkout-user-info">
								<p><strong>Account:</strong> {user?.name} ({user?.email})</p>
								<p><strong>Package:</strong> {pkg?.package_name}</p>
								<p><strong>Credits:</strong> {pkg?.credit_amount}</p>
							</div>

							{error && <div className="checkout-error">{error}</div>}

							<div className="checkout-terms">
								<label>
									<input
										type="checkbox"
										checked={agreeTerms}
										onChange={(e) => setAgreeTerms(e.target.checked)}
									/>
									<span>I agree to the <a href="#" target="_blank">Terms of Service</a> and <a href="#" target="_blank">Privacy Policy</a></span>
								</label>
							</div>

							<button
								className="checkout-continue-btn"
								onClick={handleProceedToPayment}
								disabled={!agreeTerms}
							>
								Continue to Payment
							</button>

							<div className="checkout-secure-badge">
								<span>🔒</span> Secure payment powered by Stripe
							</div>
						</div>
					)}

					{/* Processing Step (loading payment form) */}
					{step === 'processing' && !clientSecret && (
						<div className="checkout-section checkout-processing">
							<div className="processing-spinner">
								<span className="auth-spinner" />
							</div>
							<h2>Preparing Payment</h2>
							<p>Please wait while we set up your payment...</p>
						</div>
					)}

					{/* Payment Step (Stripe Elements) */}
					{step === 'payment' && clientSecret && stripePromiseState && (
						<div className="checkout-section">
							<h2>Payment Details</h2>
							{error && <div className="checkout-error">{error}</div>}

							<Elements
								stripe={stripePromiseState}
								options={{
									clientSecret,
									appearance: stripeAppearance,
								}}
							>
								<StripePaymentForm
									clientSecret={clientSecret}
									transactionId={transactionId}
									packageName={pkg?.package_name}
									credits={pkg?.credit_amount}
									onSuccess={handlePaymentSuccess}
									onError={handlePaymentError}
								/>
							</Elements>

							<div className="checkout-secure-badge">
								<span>🔒</span> Your payment information is encrypted and secure
							</div>
						</div>
					)}
				</div>
			</main>
		</div>
	);
};

export default Checkout;
