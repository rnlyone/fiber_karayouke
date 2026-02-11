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

const PaymentHistory = () => {
	const navigate = useNavigate();
	const { isAuthenticated, isLoading: authLoading } = useAuth();
	const { formatFromUSD } = useCurrency();

	const [transactions, setTransactions] = useState([]);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState(null);
	const [filter, setFilter] = useState('all'); // all, completed, pending, failed
	const [page, setPage] = useState(1);
	const [hasMore, setHasMore] = useState(false);

	const fetchTransactions = useCallback(async () => {
		try {
			setLoading(true);
			const queryParams = new URLSearchParams({
				page: page.toString(),
				limit: '10',
				...(filter !== 'all' && { status: filter }),
			});

			const response = await fetchWithAuth(`${API_BASE}/api/transactions?${queryParams}`);
			if (!response.ok) throw new Error('Failed to fetch transactions');
			
			const data = await response.json();
			
			if (page === 1) {
				setTransactions(data.transactions || data || []);
			} else {
				setTransactions((prev) => [...prev, ...(data.transactions || data || [])]);
			}
			
			setHasMore(data.has_more || false);
		} catch (err) {
			setError(err.message);
		} finally {
			setLoading(false);
		}
	}, [page, filter]);

	useEffect(() => {
		if (!authLoading && !isAuthenticated) {
			navigate('/login');
			return;
		}

		if (isAuthenticated) {
			fetchTransactions();
		}
	}, [isAuthenticated, authLoading, page, filter, navigate, fetchTransactions]);

	const getStatusBadge = (status) => {
		const badges = {
			completed: { class: 'badge-success', label: 'Completed' },
			success: { class: 'badge-success', label: 'Completed' },
			pending: { class: 'badge-pending', label: 'Pending' },
			processing: { class: 'badge-pending', label: 'Processing' },
			failed: { class: 'badge-failed', label: 'Failed' },
			cancelled: { class: 'badge-failed', label: 'Cancelled' },
			refunded: { class: 'badge-refunded', label: 'Refunded' },
		};
		return badges[status] || { class: 'badge-default', label: status };
	};

	const formatDate = (dateString) => {
		const date = new Date(dateString);
		return date.toLocaleDateString(undefined, {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
		});
	};

	const formatTime = (dateString) => {
		const date = new Date(dateString);
		return date.toLocaleTimeString(undefined, {
			hour: '2-digit',
			minute: '2-digit',
		});
	};

	const handleFilterChange = (newFilter) => {
		setFilter(newFilter);
		setPage(1);
		setTransactions([]);
	};

	if (authLoading) {
		return (
			<div className="payment-history-page">
				<div className="payment-history-loading">
					<span className="auth-spinner" />
				</div>
			</div>
		);
	}

	return (
		<div className="payment-history-page">
			<header className="payment-history-header">
				<div className="payment-history-title">
					<Link to="/dashboard" className="back-link">← Back</Link>
					<h1>Payment History</h1>
				</div>
				<Link to="/packages" className="btn-buy-credits">
					+ Buy Credits
				</Link>
			</header>

			<main className="payment-history-main">
				{/* Summary Cards */}
				<div className="payment-summary-cards">
					<div className="summary-card">
						<span className="summary-icon">💳</span>
						<div className="summary-info">
							<span className="summary-value">{transactions.length}</span>
							<span className="summary-label">Total Transactions</span>
						</div>
					</div>
					<div className="summary-card">
						<span className="summary-icon">✅</span>
						<div className="summary-info">
							<span className="summary-value">
								{transactions.filter((t) => t.status === 'completed' || t.status === 'success').length}
							</span>
							<span className="summary-label">Successful</span>
						</div>
					</div>
					<div className="summary-card">
						<span className="summary-icon">🎵</span>
						<div className="summary-info">
							<span className="summary-value">
								{transactions
									.filter((t) => t.status === 'completed' || t.status === 'success')
									.reduce((sum, t) => sum + (t.credit_amount || 0), 0)}
							</span>
							<span className="summary-label">Credits Purchased</span>
						</div>
					</div>
				</div>

				{/* Filters */}
				<div className="payment-history-filters">
					<button
						className={`filter-btn ${filter === 'all' ? 'active' : ''}`}
						onClick={() => handleFilterChange('all')}
					>
						All
					</button>
					<button
						className={`filter-btn ${filter === 'completed' ? 'active' : ''}`}
						onClick={() => handleFilterChange('completed')}
					>
						Completed
					</button>
					<button
						className={`filter-btn ${filter === 'pending' ? 'active' : ''}`}
						onClick={() => handleFilterChange('pending')}
					>
						Pending
					</button>
					<button
						className={`filter-btn ${filter === 'failed' ? 'active' : ''}`}
						onClick={() => handleFilterChange('failed')}
					>
						Failed
					</button>
				</div>

				{/* Transactions List */}
				{error ? (
					<div className="payment-history-error">
						<p>{error}</p>
						<button onClick={fetchTransactions} className="btn-retry">
							Retry
						</button>
					</div>
				) : loading && transactions.length === 0 ? (
					<div className="payment-history-loading">
						<span className="auth-spinner" />
					</div>
				) : transactions.length === 0 ? (
					<div className="payment-history-empty">
						<span className="empty-icon">📋</span>
						<h2>No Transactions Yet</h2>
						<p>Your payment history will appear here after you make your first purchase.</p>
						<Link to="/packages" className="btn-primary">
							Browse Packages
						</Link>
					</div>
				) : (
					<>
						<div className="transactions-list">
							{transactions.map((tx) => {
								const badge = getStatusBadge(tx.status);
								return (
									<div key={tx.id} className="transaction-item">
										<div className="transaction-left">
											<div className="transaction-icon">
												{tx.status === 'completed' || tx.status === 'success' ? '✅' : 
												 tx.status === 'pending' ? '⏳' : '❌'}
											</div>
											<div className="transaction-info">
												<h3>{tx.package_name || 'Credit Purchase'}</h3>
												<p className="transaction-id">ID: {tx.id}</p>
												<p className="transaction-date">
													{formatDate(tx.created_at)} at {formatTime(tx.created_at)}
												</p>
											</div>
										</div>
										<div className="transaction-right">
											<div className="transaction-amount">
												<span className="amount-main">
													{formatFromUSD((tx.amount || 0) / 100)}
												</span>
												{tx.credit_amount && (
													<span className="credits-earned">+{tx.credit_amount} credits</span>
												)}
											</div>
											<span className={`status-badge ${badge.class}`}>
												{badge.label}
											</span>
										</div>
									</div>
								);
							})}
						</div>

						{hasMore && (
							<div className="load-more-container">
								<button
									className="btn-load-more"
									onClick={() => setPage((p) => p + 1)}
									disabled={loading}
								>
									{loading ? 'Loading...' : 'Load More'}
								</button>
							</div>
						)}
					</>
				)}
			</main>
		</div>
	);
};

export default PaymentHistory;
