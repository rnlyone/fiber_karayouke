import { useEffect, useState } from 'react';
import { fetchWithAuth } from '../../lib/auth.jsx';
import AdminLayout from './AdminLayout.jsx';

const API_BASE = (() => {
	const raw = import.meta.env.VITE_WS_HOST?.trim();
	if (raw && raw.length > 0) return raw.replace(/\/$/, '');
	if (typeof window !== 'undefined' && window.location) return window.location.origin.replace(/\/$/, '');
	return '';
})();

const AdminUsers = () => {
	const [users, setUsers] = useState([]);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState(null);
	const [selectedUser, setSelectedUser] = useState(null);
	const [showAwardModal, setShowAwardModal] = useState(false);
	const [awardAmount, setAwardAmount] = useState('');
	const [awardReason, setAwardReason] = useState('');
	const [awardCreditType, setAwardCreditType] = useState('extra');
	const [awarding, setAwarding] = useState(false);
	const parsedAwardAmount = Number(awardAmount);
	const isAwardAmountValid = Number.isFinite(parsedAwardAmount) && parsedAwardAmount !== 0;

	useEffect(() => {
		fetchUsers();
	}, []);

	const fetchUsers = async () => {
		try {
			const response = await fetchWithAuth(`${API_BASE}/api/admin/users`);
			if (!response.ok) throw new Error('Failed to fetch users');
			const data = await response.json();
			setUsers(data.users || []);
		} catch (err) {
			setError(err.message);
		} finally {
			setLoading(false);
		}
	};

	const openAwardModal = (user) => {
		setSelectedUser(user);
		setAwardAmount('');
		setAwardReason('');
		setAwardCreditType('extra');
		setShowAwardModal(true);
	};

	const handleAwardCredits = async (e) => {
		e.preventDefault();
		if (!selectedUser) return;
		const amount = parsedAwardAmount;
		if (!isAwardAmountValid) {
			setError('Amount must be a non-zero number');
			return;
		}
		setAwarding(true);
		try {
			const response = await fetchWithAuth(
				`${API_BASE}/api/admin/credits/award`,
				{
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify({
						user_id: selectedUser.id,
						amount,
						credit_type: awardCreditType,
						description: awardReason || 'Admin award',
					}),
				}
			);
			if (!response.ok) throw new Error('Failed to award credits');
			await fetchUsers();
			setShowAwardModal(false);
		} catch (err) {
			setError(err.message);
		} finally {
			setAwarding(false);
		}
	};

	return (
		<AdminLayout title="Users">
			{loading ? (
				<div className="admin-loading-content">
					<span className="auth-spinner" />
				</div>
			) : (
				<div className="admin-users">
					{error && <div className="admin-error">{error}</div>}

					<div className="admin-table-container">
						<table className="admin-table">
							<thead>
								<tr>
									<th>User</th>
									<th>Email</th>
									<th>Free / Extra Credits</th>
									<th>Plan</th>
									<th>Rooms</th>
									<th>Joined</th>
									<th>Actions</th>
								</tr>
							</thead>
							<tbody>
								{users.length === 0 ? (
									<tr>
										<td colSpan="7" className="admin-table-empty">
											No users found.
										</td>
									</tr>
								) : (
									users.map((user) => (
										<tr key={user.id}>
											<td>
												<div className="admin-cell-primary">{user.name}</div>
												<div className="admin-cell-secondary">@{user.username}</div>
											</td>
											<td>{user.email}</td>
											<td>
												<span className="admin-credits-badge">🎁 {user.free_credit ?? 0}</span>
												{' / '}
												<span className="admin-credits-badge">💎 {user.extra_credit ?? user.credit ?? 0}</span>
											</td>
											<td>
												{user.subscription_plan_name ? (
													<span className="admin-badge success">{user.subscription_plan_name}</span>
												) : (
													<span className="admin-badge muted">Free</span>
												)}
											</td>
											<td>{user.roomCount || 0}</td>
											<td>
												{user.createdAt || user.created_at
													? new Date(user.createdAt || user.created_at).toLocaleDateString('id-ID', {
															day: 'numeric',
															month: 'short',
															year: 'numeric',
													  })
													: '—'}
											</td>
											<td>
												<div className="admin-actions">
													<button
														className="admin-btn admin-btn-sm admin-btn-success"
														onClick={() => openAwardModal(user)}
													>
														Award Credits
													</button>
												</div>
											</td>
										</tr>
									))
								)}
							</tbody>
						</table>
					</div>
				</div>
			)}

			{showAwardModal && selectedUser && (
				<div className="admin-modal-overlay" onClick={() => setShowAwardModal(false)}>
					<div className="admin-modal" onClick={(e) => e.stopPropagation()}>
						<div className="admin-modal-header">
							<h3>Award Credits</h3>
							<button className="admin-modal-close" onClick={() => setShowAwardModal(false)}>
								×
							</button>
						</div>
						<form onSubmit={handleAwardCredits} className="admin-modal-body">
							<div className="admin-award-user">
								<div className="admin-award-user-info">
									<strong>{selectedUser.name}</strong>
									<span>{selectedUser.email}</span>
								</div>
								<div className="admin-award-current">
									🎁 Free: <strong>{selectedUser.free_credit ?? 0}</strong> &nbsp;|&nbsp; 💎 Extra: <strong>{selectedUser.extra_credit ?? selectedUser.credit ?? 0}</strong>
								</div>
							</div>
							<div className="admin-form-group">
								<label>Credit Type</label>
								<select
									className="admin-input"
									value={awardCreditType}
									onChange={(e) => setAwardCreditType(e.target.value)}
								>
									<option value="extra">💎 Extra Credit</option>
									<option value="free">🎁 Free Credit</option>
								</select>
							</div>
							<div className="admin-form-group">
								<label>Amount (can be negative to deduct)</label>
								<input
									type="number"
									className="admin-input"
									value={awardAmount}
									onChange={(e) => setAwardAmount(e.target.value)}
									required
									placeholder="e.g., 10 or -5"
								/>
							</div>
							<div className="admin-form-group">
								<label>Reason (optional)</label>
								<input
									type="text"
									className="admin-input"
									value={awardReason}
									onChange={(e) => setAwardReason(e.target.value)}
									placeholder="e.g., Promotional bonus"
								/>
							</div>
							<div className="admin-modal-footer">
								<button
									type="button"
									className="admin-btn"
									onClick={() => setShowAwardModal(false)}
								>
									Cancel
								</button>
								<button
									type="submit"
									className="admin-btn admin-btn-primary"
									disabled={awarding || !isAwardAmountValid}
								>
									{awarding ? 'Awarding...' : 'Award Credits'}
								</button>
							</div>
						</form>
					</div>
				</div>
			)}
		</AdminLayout>
	);
};

export default AdminUsers;
