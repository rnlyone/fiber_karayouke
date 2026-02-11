import { useEffect, useState } from 'react';
import { fetchWithAuth } from '../../lib/auth.jsx';
import { formatIDR } from '../../lib/currency.jsx';
import AdminLayout from './AdminLayout.jsx';

const API_BASE = (() => {
	const raw = import.meta.env.VITE_WS_HOST?.trim();
	if (raw && raw.length > 0) return raw.replace(/\/$/, '');
	if (typeof window !== 'undefined' && window.location) return window.location.origin.replace(/\/$/, '');
	return '';
})();

const statusColors = {
	pending: 'warning',
	settlement: 'success',
	failed: 'danger',
	expired: 'muted',
	refunded: 'info',
};

const AdminTransactions = () => {
	const [transactions, setTransactions] = useState([]);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState(null);
	const [updating, setUpdating] = useState({});

	useEffect(() => {
		fetchTransactions();
	}, []);

	const fetchTransactions = async () => {
		try {
			const response = await fetchWithAuth(`${API_BASE}/api/admin/transactions`);
			if (!response.ok) throw new Error('Failed to fetch transactions');
			const data = await response.json();
			setTransactions(data.transactions || []);
		} catch (err) {
			setError(err.message);
		} finally {
			setLoading(false);
		}
	};

	const handleStatusChange = async (id, newStatus) => {
		if (!confirm(`Are you sure you want to change status to "${newStatus}"?`)) return;
		setUpdating((prev) => ({ ...prev, [id]: true }));
		try {
			const response = await fetchWithAuth(`${API_BASE}/api/admin/transactions/${id}/status`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ status: newStatus }),
			});
			if (!response.ok) throw new Error('Failed to update status');
			await fetchTransactions();
		} catch (err) {
			setError(err.message);
		} finally {
			setUpdating((prev) => ({ ...prev, [id]: false }));
		}
	};

	return (
		<AdminLayout title="Transactions">
			{loading ? (
				<div className="admin-loading-content">
					<span className="auth-spinner" />
				</div>
			) : (
				<div className="admin-transactions">
					{error && <div className="admin-error">{error}</div>}

					<div className="admin-table-container">
						<table className="admin-table">
							<thead>
								<tr>
									<th>ID</th>
									<th>User</th>
									<th>Package / Plan</th>
									<th>Amount (IDR)</th>
									<th>Credits</th>
									<th>Status</th>
									<th>Date</th>
									<th>Actions</th>
								</tr>
							</thead>
							<tbody>
								{transactions.length === 0 ? (
									<tr>
										<td colSpan="8" className="admin-table-empty">
											No transactions found.
										</td>
									</tr>
								) : (
									transactions.map((tx) => (
										<tr key={tx.id}>
											<td className="admin-cell-mono">
												{tx.id.toString().padStart(6, '0')}
											</td>
											<td>
												<div className="admin-cell-primary">{tx.user?.name}</div>
												<div className="admin-cell-secondary">{tx.user?.email}</div>
											</td>
											<td>
												<div className="admin-cell-primary">{tx.package?.package_name || tx.plan?.plan_name || '-'}</div>
												<div className="admin-cell-secondary">{tx.tx_type === 'subscription' ? '📋 Subscription' : '💎 Extra Credit'}</div>
											</td>
											<td>{formatIDR(tx.amount)}</td>
											<td>{tx.credits}</td>
											<td>
												<span className={`admin-badge ${statusColors[tx.status]}`}>
													{tx.status}
												</span>
											</td>
											<td>
												{new Date(tx.createdAt).toLocaleDateString('id-ID', {
													day: 'numeric',
													month: 'short',
													year: 'numeric',
													hour: '2-digit',
													minute: '2-digit',
												})}
											</td>
											<td>
												{tx.status === 'pending' && (
													<div className="admin-actions">
														<button
															className="admin-btn admin-btn-sm admin-btn-success"
															onClick={() => handleStatusChange(tx.id, 'settlement')}
															disabled={updating[tx.id]}
														>
															Settle
														</button>
														<button
															className="admin-btn admin-btn-sm admin-btn-danger"
															onClick={() => handleStatusChange(tx.id, 'failed')}
															disabled={updating[tx.id]}
														>
															Fail
														</button>
													</div>
												)}
												{tx.status === 'settlement' && (
													<span className="admin-cell-secondary">Completed</span>
												)}
												{tx.status === 'failed' && (
													<span className="admin-cell-secondary">-</span>
												)}
											</td>
										</tr>
									))
								)}
							</tbody>
						</table>
					</div>
				</div>
			)}
		</AdminLayout>
	);
};

export default AdminTransactions;
