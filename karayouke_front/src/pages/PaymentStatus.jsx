import { useState, useEffect, useCallback, useMemo } from 'react';
import { Link, useParams, useLocation, useNavigate } from 'react-router-dom';
import { useAuth, fetchWithAuth, getAuthToken } from '../lib/auth.jsx';
import { useCurrency } from '../lib/currency.jsx';

const API_BASE = (() => {
	const raw = import.meta.env.VITE_WS_HOST?.trim();
	if (raw && raw.length > 0) return raw.replace(/\/$/, '');
	if (typeof window !== 'undefined' && window.location) return window.location.origin.replace(/\/$/, '');
	return '';
})();

const PaymentStatus = () => {
	const { transactionId } = useParams();
	const location = useLocation();
	const navigate = useNavigate();
	const { isAuthenticated, isLoading: authLoading } = useAuth();
	const { formatFromUSD } = useCurrency();

	const [transaction, setTransaction] = useState(null);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState(null);
	const [pollCount, setPollCount] = useState(0);
	const [confirming, setConfirming] = useState(false);
	const [stripeStatus, setStripeStatus] = useState(null);

	// Get initial state from navigation
	const initialState = location.state || {};

	// Parse Stripe redirect parameters from URL
	const stripeParams = useMemo(() => {
		const searchParams = new URLSearchParams(location.search);
		return {
			paymentIntent: searchParams.get('payment_intent'),
			clientSecret: searchParams.get('payment_intent_client_secret'),
			redirectStatus: searchParams.get('redirect_status'),
		};
	}, [location.search]);

	// Check if this is a redirect from Stripe
	const isStripeRedirect = !!stripeParams.redirectStatus;

	const fetchTransaction = useCallback(async () => {
		try {
			const response = await fetchWithAuth(`${API_BASE}/api/transactions/${transactionId}`);
			if (!response.ok) {
				if (response.status === 404) {
					throw new Error('Transaction not found');
				}
				// Handle auth error - don't redirect if we have a token
				if (response.status === 401 && getAuthToken()) {
					// Token might be expired, retry once
					throw new Error('Session expired. Please log in again.');
				}
				throw new Error('Failed to fetch transaction status');
			}
			const data = await response.json();
			setTransaction(data);
		} catch (err) {
			setError(err.message);
		} finally {
			setLoading(false);
		}
	}, [transactionId]);

	const confirmStripePayment = useCallback(async (tx, paymentIntentId = null) => {
		const piId = paymentIntentId || tx?.external_id;
		if (!piId || confirming) return null;
		setConfirming(true);
		try {
			const response = await fetchWithAuth(`${API_BASE}/api/stripe/confirm-payment`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					payment_intent_id: piId,
					transaction_id: tx?.id || transactionId,
				}),
			});
			if (response.ok) {
				const data = await response.json();
				return data;
			}
			return null;
		} catch {
			// ignore - status will be rechecked on polling
			return null;
		} finally {
			setConfirming(false);
		}
	}, [confirming, transactionId]);

	useEffect(() => {
		// Don't redirect if we have a token (might be from Stripe redirect)
		// The token should still be in localStorage even after redirect
		const hasToken = !!getAuthToken();
		
		if (!authLoading && !isAuthenticated && !hasToken) {
			// Save current URL for redirect back after login
			navigate('/login', { state: { from: location.pathname + location.search } });
			return;
		}

		// Handle Stripe redirect - check actual payment status
		if ((isAuthenticated || hasToken) && transactionId && isStripeRedirect) {
			const handleStripeRedirect = async () => {
				// Map Stripe redirect_status to our status
				// Stripe redirect_status values: 'succeeded', 'processing', 'requires_payment_method', 'failed'
				setStripeStatus(stripeParams.redirectStatus);
				
				// Confirm with backend to get actual status
				if (stripeParams.paymentIntent) {
					const result = await confirmStripePayment(null, stripeParams.paymentIntent);
					if (result) {
						// Update stripe status based on backend response
						if (result.paymentStatus) {
							setStripeStatus(result.paymentStatus);
						}
					}
				}
				
				fetchTransaction();
			};
			handleStripeRedirect();
		} else if ((isAuthenticated || hasToken) && transactionId) {
			fetchTransaction();
		}
	}, [transactionId, isAuthenticated, authLoading, navigate, fetchTransaction, isStripeRedirect, stripeParams, confirmStripePayment, location]);

	// Poll for status updates if payment is pending
	useEffect(() => {
		if (transaction && transaction.status === 'pending' && pollCount < 30) {
			void confirmStripePayment(transaction);
			const timer = setTimeout(() => {
				fetchTransaction();
				setPollCount((c) => c + 1);
			}, 3000);
			return () => clearTimeout(timer);
		}
	}, [transaction, pollCount, fetchTransaction, confirmStripePayment]);

	const getStatusIcon = (status) => {
		switch (status) {
			case 'completed':
			case 'success':
				return '✅';
			case 'pending':
			case 'processing':
				return '⏳';
			case 'failed':
			case 'cancelled':
				return '❌';
			default:
				return '📋';
		}
	};

	const getStatusTitle = (status) => {
		switch (status) {
			case 'completed':
			case 'success':
				return 'Payment Successful!';
			case 'pending':
				return 'Payment Pending';
			case 'processing':
				return 'Processing Payment';
			case 'failed':
				return 'Payment Failed';
			case 'cancelled':
				return 'Payment Cancelled';
			default:
				return 'Payment Status';
		}
	};

	const getStatusMessage = (status) => {
		switch (status) {
			case 'completed':
			case 'success':
				return 'Your payment has been processed successfully. Credits have been added to your account.';
			case 'pending':
				return 'Your payment is being processed. This may take a few minutes.';
			case 'processing':
				return 'We are currently processing your payment. Please wait...';
			case 'failed':
				return 'Your payment could not be processed. Please try again or use a different payment method.';
			case 'cancelled':
				return 'This payment has been cancelled.';
			default:
				return 'Checking payment status...';
		}
	};

	const getStatusClass = (status) => {
		switch (status) {
			case 'completed':
			case 'success':
				return 'status-success';
			case 'pending':
			case 'processing':
				return 'status-pending';
			case 'failed':
			case 'cancelled':
				return 'status-failed';
			default:
				return '';
		}
	};

	if (authLoading || loading) {
		return (
			<div className="payment-status-page">
				<div className="payment-status-loading">
					<span className="auth-spinner" />
					<p>Checking payment status...</p>
				</div>
			</div>
		);
	}

	if (error) {
		return (
			<div className="payment-status-page">
				<div className="payment-status-content">
					<div className="payment-status-icon status-failed">❌</div>
					<h1>Error</h1>
					<p>{error}</p>
					<div className="payment-status-actions">
						<Link to="/packages" className="btn-primary">
							Browse Packages
						</Link>
						<Link to="/dashboard" className="btn-secondary">
							Go to Dashboard
						</Link>
					</div>
				</div>
			</div>
		);
	}

	// Determine the effective status based on all available information
	const determineStatus = () => {
		// If transaction is completed/failed, that's authoritative
		if (transaction?.status === 'completed' || transaction?.status === 'settlement') {
			return 'success';
		}
		if (transaction?.status === 'failed' || transaction?.status === 'cancelled') {
			return 'failed';
		}
		
		// Check Stripe redirect status for immediate feedback
		if (isStripeRedirect) {
			// Use stripeStatus if available (from confirm-payment), otherwise use redirect param
			const effectiveStripeStatus = stripeStatus || stripeParams.redirectStatus;
			
			switch (effectiveStripeStatus) {
				case 'succeeded':
					return 'success';
				case 'processing':
					return 'processing';
				case 'requires_payment_method':
				case 'requires_action':
					return 'failed'; // Payment method failed, user needs to try again
				case 'canceled':
					return 'cancelled';
				default:
					// For unknown status, check if we're still pending
					if (transaction?.status === 'pending') {
						return 'pending';
					}
			}
		}
		
		// Use transaction status or initial state
		if (transaction?.status) {
			return transaction.status;
		}
		
		return initialState.success ? 'success' : 'pending';
	};
	
	const status = determineStatus();
	const packageName = transaction?.package_name || initialState.packageName;
	const credits = transaction?.credit_amount || initialState.credits;
	const amount = transaction?.amount || 0;

	return (
		<div className="payment-status-page">
			<div className="payment-status-content">
				<div className={`payment-status-icon ${getStatusClass(status)}`}>
					{getStatusIcon(status)}
				</div>

				<h1>{getStatusTitle(status)}</h1>
				<p className="payment-status-message">{getStatusMessage(status)}</p>

				{transaction && (
					<div className="payment-status-details">
						<h2>Transaction Details</h2>
						<div className="payment-detail-row">
							<span>Transaction ID</span>
							<span className="payment-detail-value">{transaction.id}</span>
						</div>
						{packageName && (
							<div className="payment-detail-row">
								<span>Package</span>
								<span className="payment-detail-value">{packageName}</span>
							</div>
						)}
						{credits && (
							<div className="payment-detail-row">
								<span>Credits</span>
								<span className="payment-detail-value">{credits}</span>
							</div>
						)}
						{amount > 0 && (
							<div className="payment-detail-row">
								<span>Amount</span>
								<span className="payment-detail-value">
									{formatFromUSD(amount / 100)}
								</span>
							</div>
						)}
						{transaction.payment_method && (
							<div className="payment-detail-row">
								<span>Payment Method</span>
								<span className="payment-detail-value" style={{ textTransform: 'capitalize' }}>
									{transaction.payment_method}
								</span>
							</div>
						)}
						{transaction.created_at && (
							<div className="payment-detail-row">
								<span>Date</span>
								<span className="payment-detail-value">
									{new Date(transaction.created_at).toLocaleString()}
								</span>
							</div>
						)}
					</div>
				)}

				{status === 'pending' && (
					<div className="payment-status-pending-info">
						<p>
							<span className="auth-spinner" style={{ marginRight: '10px' }} />
							Checking for updates...
						</p>
					</div>
				)}

				{(status === 'completed' || status === 'success') && (
					<div className="payment-status-success-info">
						<p>🎉 Your credits are now available!</p>
					</div>
				)}

				{(status === 'failed' || status === 'cancelled') && (
					<div className="payment-status-failed-info">
						<p>Need help? <a href="mailto:support@karayouke.com">Contact Support</a></p>
					</div>
				)}

				<div className="payment-status-actions">
					{(status === 'completed' || status === 'success') && (
						<>
							<Link to="/dashboard" className="btn-primary">
								Go to Dashboard
							</Link>
							<Link to="/payment/history" className="btn-secondary">
								View Payment History
							</Link>
						</>
					)}
					{status === 'pending' && (
						<Link to="/dashboard" className="btn-secondary">
							Go to Dashboard
						</Link>
					)}
					{(status === 'failed' || status === 'cancelled') && (
						<>
							<Link to="/packages" className="btn-primary">
								Try Again
							</Link>
							<Link to="/dashboard" className="btn-secondary">
								Go to Dashboard
							</Link>
						</>
					)}
				</div>
			</div>
		</div>
	);
};

export default PaymentStatus;
