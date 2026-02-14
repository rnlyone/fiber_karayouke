import { useEffect, useState } from 'react';
import { fetchWithAuth } from '../../lib/auth.jsx';
import AdminLayout from './AdminLayout.jsx';

const API_BASE = (() => {
	const raw = import.meta.env.VITE_WS_HOST?.trim();
	if (raw && raw.length > 0) return raw.replace(/\/$/, '');
	if (typeof window !== 'undefined' && window.location) return window.location.origin.replace(/\/$/, '');
	return '';
})();

const AdminRooms = () => {
	const [rooms, setRooms] = useState([]);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState(null);
	const [filter, setFilter] = useState('all'); // all, active, expired

	useEffect(() => {
		fetchRooms();
	}, []);

	const fetchRooms = async () => {
		try {
			const response = await fetchWithAuth(`${API_BASE}/api/admin/rooms`);
			if (!response.ok) throw new Error('Failed to fetch rooms');
			const data = await response.json();
			setRooms(data.rooms || []);
		} catch (err) {
			setError(err.message);
		} finally {
			setLoading(false);
		}
	};

	const filteredRooms = rooms.filter((room) => {
		if (filter === 'all') return true;
		if (filter === 'active') return !room.is_expired;
		if (filter === 'expired') return room.is_expired;
		return true;
	});

	const formatTimeRemaining = (expiredAt, isExpired) => {
		if (!expiredAt) return 'No expiry';
		const expDate = new Date(expiredAt);
		const now = new Date();
		const diff = expDate - now;
		const absDiff = Math.abs(diff);
		const minutes = Math.floor(absDiff / 60000);
		const hours = Math.floor(minutes / 60);
		const remainMinutes = minutes % 60;

		if (isExpired || diff <= 0) {
			if (hours >= 24) {
				const days = Math.floor(hours / 24);
				return `Expired ${days}d ago`;
			}
			if (hours > 0) return `Expired ${hours}h ${remainMinutes}m ago`;
			return `Expired ${minutes}m ago`;
		}

		if (hours > 0) return `${hours}h ${remainMinutes}m remaining`;
		return `${minutes}m remaining`;
	};

	const formatExpiry = (expiredAt) => {
		if (!expiredAt) return '-';
		return new Date(expiredAt).toLocaleDateString('id-ID', {
			day: 'numeric',
			month: 'short',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit',
		});
	};

	return (
		<AdminLayout title="Rooms">
			{loading ? (
				<div className="admin-loading-content">
					<span className="auth-spinner" />
				</div>
			) : (
				<div className="admin-rooms">
					<div className="admin-toolbar">
						<div className="admin-filter-tabs">
							<button
								className={`admin-filter-tab ${filter === 'all' ? 'active' : ''}`}
								onClick={() => setFilter('all')}
							>
								All ({rooms.length})
							</button>
							<button
								className={`admin-filter-tab ${filter === 'active' ? 'active' : ''}`}
								onClick={() => setFilter('active')}
							>
								Active ({rooms.filter((r) => !r.is_expired).length})
							</button>
							<button
								className={`admin-filter-tab ${filter === 'expired' ? 'active' : ''}`}
								onClick={() => setFilter('expired')}
							>
								Expired ({rooms.filter((r) => r.is_expired).length})
							</button>
						</div>
					</div>

					{error && <div className="admin-error">{error}</div>}

					<div className="admin-table-container">
						<table className="admin-table">
							<thead>
								<tr>
									<th>Room</th>
									<th>Owner</th>
									<th>Code</th>
									<th>Status</th>
									<th>Time Remaining</th>
									<th>Created</th>
									<th>Expires</th>
								</tr>
							</thead>
							<tbody>
								{filteredRooms.length === 0 ? (
									<tr>
										<td colSpan="7" className="admin-table-empty">
											No rooms found.
										</td>
									</tr>
								) : (
									filteredRooms.map((room) => (
										<tr key={room.id} className={room.is_expired ? 'expired-row' : ''}>
											<td>
												<div className="admin-cell-primary">{room.room_name}</div>
											</td>
											<td>
												<div className="admin-cell-primary">
													{room.creator?.name || '-'}
												</div>
												<div className="admin-cell-secondary">
													{room.creator?.email}
												</div>
											</td>
											<td>
												<code className="admin-room-code">{room.room_key}</code>
											</td>
											<td>
												<span
													className={`admin-badge ${room.is_expired ? 'danger' : 'success'}`}
												>
													{room.is_expired ? '● Expired' : '● Active'}
												</span>
											</td>
											<td>
												<span
													className={`admin-time-remaining ${room.is_expired ? 'expired' : ''}`}
												>
													{formatTimeRemaining(room.expired_at, room.is_expired)}
												</span>
											</td>
											<td>
												{new Date(room.created_at).toLocaleDateString('id-ID', {
													day: 'numeric',
													month: 'short',
													year: 'numeric',
													hour: '2-digit',
													minute: '2-digit',
												})}
											</td>
											<td>
												<div className="admin-cell-secondary">
													{formatExpiry(room.expired_at)}
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
		</AdminLayout>
	);
};

export default AdminRooms;
