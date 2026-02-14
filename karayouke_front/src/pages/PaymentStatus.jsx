import { useState, useEffect, useCallback, useRef } from 'react';
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
	const { format } = useCurrency();

	const [transaction, setTransaction] = useState(null);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState(null);
	const [pollCount, setPollCount] = useState(0);
	const [popupOpen, setPopupOpen] = useState(false);
	const popupTriggered = useRef(false);

	// Get initial state from navigation (from Packages page)
	const navState = location.state || {};

	const fetchTransaction = useCallback(async () => {
		try {
			const response = await fetchWithAuth(`${API_BASE}/api/transactions/${transactionId}`);
			if (!response.ok) {
				if (response.status === 404) throw new Error('Transaction not found');
				if (response.status === 401 && getAuthToken()) {
					throw new Error('Session expired. Please log in again.');
				}
				throw new Error('Failed to fetch transaction status');
			}
			const data = await response.json();
			setTransaction(data);
			return data;
		} catch (err) {
			setError(err.message);
			return null;
		} finally {
			setLoading(false);
		}
	}, [transactionId]);

	// Open the Flip payment popup or redirect to link
	const openFlipPayment = useCallback((companyCode, productCode, linkUrl) => {
		if (companyCode && productCode && window.FlipCheckout) {
			setPopupOpen(true);
			window.FlipCheckout.pay(companyCode, productCode, {
				onSuccess: () => { setPopupOpen(false); fetchTransaction(); },
				onPending: () => { setPopupOpen(false); fetchTransaction(); },
				onClose: () => { setPopupOpen(false); },
			});
		} else if (linkUrl) {
			window.open(linkUrl, '_blank');
		}
	}, [fetchTransaction]);

	useEffect(() => {
		const hasToken = !!getAuthToken();
		if (!authLoading && !isAuthenticated && !hasToken) {
			navigate('/login', { state: { from: location.pathname } });
			return;
		}
		if (isAuthenticated || hasToken) {
			fetchTransaction().then((tx) => {
				// Auto-open popup if navigated from Packages page with autoOpen flag
				if (navState.autoOpen && !popupTriggered.current && tx && tx.status === 'pending') {
					popupTriggered.current = true;
					const cc = navState.companyCode || tx.flip_company_code;
					const pc = navState.productCode || tx.flip_product_code;
					const url = navState.linkUrl || tx.flip_url;
					if (cc || url) {
						setTimeout(() => openFlipPayment(cc, pc, url), 500);
					}
				}
			});
		}
	}, [transactionId, isAuthenticated, authLoading]); // eslint-disable-line react-hooks/exhaustive-deps

	// Poll for status updates if payment is pending (pause while popup is open)
	useEffect(() => {
		if (transaction && transaction.status === 'pending' && pollCount < 60 && !popupOpen) {
			const timer = setTimeout(() => {
				fetchTransaction();
				setPollCount((c) => c + 1);
			}, 5000);
			return () => clearTimeout(timer);
		}
	}, [transaction, pollCount, fetchTransaction, popupOpen]);

	// Handle "Continue Payment" button click
	const handleContinuePayment = () => {
		if (!transaction) return;
		openFlipPayment(
			transaction.flip_company_code,
			transaction.flip_product_code,
			transaction.flip_url,
		);
	};

	const getStatusIcon = (status) => {
		switch (status) {
			case 'settlement': return '✅';
			case 'pending': return '⏳';
			case 'failed':
			case 'expired':
			case 'refunded': return '❌';
			default: return '📋';
		}
	};

	const getStatusTitle = (status) => {
		switch (status) {
			case 'settlement': return 'Payment Successful!';
			case 'pending': return 'Awaiting Payment';
			case 'failed': return 'Payment Failed';
			case 'expired': return 'Payment Expired';
			case 'refunded': return 'Payment Refunded';
			default: return 'Payment Status';
		}
	};

	const getStatusMessage = (status, txType) => {
		const isSubscription = txType === 'subscription';
		switch (status) {
			case 'settlement':
				return isSubscription
					? 'Your subscription has been activated! Enjoy your upgraded plan.'
					: 'Your credits have been added to your account.';
			case 'pending':
				return 'Please complete your payment. Click the button below to open the payment page.';
			case 'failed':
				return 'Your payment could not be processed. Please try again.';
			case 'expired':
				return 'This payment has expired. Please create a new order.';
			case 'refunded':
				return 'This payment has been refunded.';
			default:
				return 'Checking payment status...';
		}
	};

	const getStatusClass = (status) => {
		switch (status) {
			case 'settlement': return 'status-success';
			case 'pending': return 'status-pending';
			case 'failed':
			case 'expired':
			case 'refunded': return 'status-failed';
			default: return '';
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
						<Link to="/packages" className="btn-primary">Browse Packages</Link>
						<Link to="/" className="btn-secondary">Go to Dashboard</Link>
					</div>
				</div>
			</div>
		);
	}

	const status = transaction?.status || 'pending';
	const txType = transaction?.tx_type || 'extra_credit';
	const packageName = transaction?.item_name || navState.packageName;
	const credits = transaction?.credit_amount;
	const amount = transaction?.amount || 0;
	const hasPaymentLink = !!(transaction?.flip_company_code || transaction?.flip_url);

	return (
		<div className="payment-status-page">
			<div className="payment-status-content">
				<div className={`payment-status-icon ${getStatusClass(status)}`}>
					{getStatusIcon(status)}
				</div>

				<h1>{getStatusTitle(status)}</h1>
				<p className="payment-status-message">{getStatusMessage(status, txType)}</p>

				{transaction && (
					<div className="payment-status-details">
						<h2>Transaction Details</h2>
						<div className="payment-detail-row">
							<span>Transaction ID</span>
							<span className="payment-detail-value">{transaction.id}</span>
						</div>
						<div className="payment-detail-row">
							<span>Type</span>
							<span className="payment-detail-value" style={{ textTransform: 'capitalize' }}>
								{txType === 'subscription' ? '📋 Subscription' : '💎 Extra Credits'}
							</span>
						</div>
						{packageName && (
							<div className="payment-detail-row">
								<span>{txType === 'subscription' ? 'Plan' : 'Package'}</span>
								<span className="payment-detail-value">{packageName}</span>
							</div>
						)}
						{credits > 0 && txType !== 'subscription' && (
							<div className="payment-detail-row">
								<span>Credits</span>
								<span className="payment-detail-value">{credits}</span>
							</div>
						)}
						{amount > 0 && (
							<div className="payment-detail-row">
								<span>Amount</span>
								<span className="payment-detail-value">{format(amount)}</span>
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

				{status === 'pending' && hasPaymentLink && (
					<div className="payment-status-pending-info">
						<button className="btn-primary btn-pay" onClick={handleContinuePayment} disabled={popupOpen}>
							{popupOpen ? (
								<><span className="auth-spinner" style={{ marginRight: '10px' }} /> Payment in progress...</>
							) : (
								'💳 Continue Payment'
							)}
						</button>
						<p style={{ marginTop: '16px', opacity: 0.7 }}>
							<span className="auth-spinner" style={{ marginRight: '8px' }} />
							Waiting for payment confirmation... ({pollCount})
						</p>
					</div>
				)}

				{status === 'pending' && !hasPaymentLink && (
					<div className="payment-status-pending-info">
						<p>
							<span className="auth-spinner" style={{ marginRight: '10px' }} />
							Checking for updates... ({pollCount})
						</p>
					</div>
				)}

				{status === 'settlement' && (
					<div className="payment-status-success-info">
						<p>🎉 {txType === 'subscription' ? 'Your subscription is now active!' : 'Your credits are now available!'}</p>
					</div>
				)}

				{(status === 'failed' || status === 'expired') && (
					<div className="payment-status-failed-info">
						<p>Need help? <a href="mailto:ask@karayouke.com">Contact Support</a></p>
					</div>
				)}

				<div className="payment-status-actions">
					{status === 'settlement' && (
						<>
							<Link to="/" className="btn-primary">Go to Dashboard</Link>
							<Link to="/payment/history" className="btn-secondary">View Payment History</Link>
						</>
					)}
					{status === 'pending' && (
						<Link to="/packages" className="btn-secondary">← Back to Packages</Link>
					)}
					{(status === 'failed' || status === 'expired') && (
						<>
							<Link to="/packages" className="btn-primary">Try Again</Link>
							<Link to="/" className="btn-secondary">Go to Dashboard</Link>
						</>
					)}
				</div>
			</div>
		</div>
	);
};

export default PaymentStatus;
