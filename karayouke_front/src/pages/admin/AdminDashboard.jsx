import { useEffect, useState } from 'react';
import { fetchWithAuth } from '../../lib/auth.jsx';
import AdminLayout from './AdminLayout.jsx';

const API_BASE = (() => {
	const raw = import.meta.env.VITE_WS_HOST?.trim();
	if (raw && raw.length > 0) return raw.replace(/\/$/, '');
	if (typeof window !== 'undefined' && window.location) return window.location.origin.replace(/\/$/, '');
	return '';
})();

const AdminDashboard = () => {
	const [stats, setStats] = useState(null);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState(null);

	const formatNumber = (value) => (value ?? 0).toLocaleString('en-US');
	const formatCurrency = (cents) => {
		const amount = (cents ?? 0) / 100;
		return amount.toLocaleString('en-US', { style: 'currency', currency: 'USD' });
	};

	const renderBars = (values = [], colorClass = '') => {
		const maxValue = Math.max(1, ...values);
		return (
			<div className={`admin-chart-bars ${colorClass}`}>
				{values.map((value, index) => (
					<div className="admin-chart-bar" key={`bar-${index}`}>
						<span className="admin-chart-bar-fill" style={{ height: `${(value / maxValue) * 100}%` }} />
						<span className="admin-chart-bar-value">{value}</span>
					</div>
				))}
			</div>
		);
	};

	useEffect(() => {
		const fetchStats = async () => {
			try {
				const response = await fetchWithAuth(`${API_BASE}/api/admin/dashboard`);
				if (!response.ok) throw new Error('Failed to fetch dashboard stats');
				const data = await response.json();
				setStats(data);
			} catch (err) {
				setError(err.message);
			} finally {
				setLoading(false);
			}
		};
		fetchStats();
	}, []);

	return (
		<AdminLayout title="Dashboard">
			{loading ? (
				<div className="admin-loading-content">
					<span className="auth-spinner" />
				</div>
			) : error ? (
				<div className="admin-error">{error}</div>
			) : (
				<div className="admin-dashboard">
					<div className="admin-stats-grid">
						<div className="admin-stat-card">
							<div className="admin-stat-icon">👥</div>
							<div className="admin-stat-content">
								<span className="admin-stat-value">{formatNumber(stats?.totalUsers)}</span>
								<span className="admin-stat-label">Total Users</span>
							</div>
						</div>
						<div className="admin-stat-card">
							<div className="admin-stat-icon">🎤</div>
							<div className="admin-stat-content">
								<span className="admin-stat-value">{formatNumber(stats?.totalRooms)}</span>
								<span className="admin-stat-label">Total Rooms</span>
							</div>
						</div>
						<div className="admin-stat-card active">
							<div className="admin-stat-icon">🟢</div>
							<div className="admin-stat-content">
								<span className="admin-stat-value">{formatNumber(stats?.activeRooms)}</span>
								<span className="admin-stat-label">Active Rooms</span>
							</div>
						</div>
						<div className="admin-stat-card expired">
							<div className="admin-stat-icon">🔴</div>
							<div className="admin-stat-content">
								<span className="admin-stat-value">{formatNumber(stats?.expiredRooms)}</span>
								<span className="admin-stat-label">Expired Rooms</span>
							</div>
						</div>
						<div className="admin-stat-card">
							<div className="admin-stat-icon">📦</div>
							<div className="admin-stat-content">
								<span className="admin-stat-value">{formatNumber(stats?.totalPackages)}</span>
								<span className="admin-stat-label">Packages</span>
							</div>
						</div>
						<div className="admin-stat-card pending">
							<div className="admin-stat-icon">⏳</div>
							<div className="admin-stat-content">
								<span className="admin-stat-value">{formatNumber(stats?.pendingTransactions)}</span>
								<span className="admin-stat-label">Pending Transactions</span>
							</div>
						</div>
						<div className="admin-stat-card">
							<div className="admin-stat-icon">✅</div>
							<div className="admin-stat-content">
								<span className="admin-stat-value">{formatNumber(stats?.settledTransactions)}</span>
								<span className="admin-stat-label">Settled Transactions</span>
							</div>
						</div>
						<div className="admin-stat-card">
							<div className="admin-stat-icon">❌</div>
							<div className="admin-stat-content">
								<span className="admin-stat-value">{formatNumber(stats?.failedTransactions)}</span>
								<span className="admin-stat-label">Failed Transactions</span>
							</div>
						</div>
					</div>

					<div className="admin-dashboard-grid">
						<div className="admin-card">
							<h3 className="admin-card-title">Business Highlights</h3>
							<p className="admin-card-subtitle">Key revenue and engagement indicators</p>
							<div className="admin-highlight-grid">
								<div className="admin-highlight">
									<span className="admin-highlight-label">Total Revenue</span>
									<span className="admin-highlight-value">{formatCurrency(stats?.totalRevenue)}</span>
								</div>
								<div className="admin-highlight">
									<span className="admin-highlight-label">Credits Awarded</span>
									<span className="admin-highlight-value">{formatNumber(stats?.totalCreditsAwarded)}</span>
								</div>
								<div className="admin-highlight">
									<span className="admin-highlight-label">Total Transactions</span>
									<span className="admin-highlight-value">{formatNumber(stats?.totalTransactions)}</span>
								</div>
							</div>
							<div className="admin-progress">
								<div>
									<span className="admin-progress-label">Active Room Rate</span>
									<span className="admin-progress-value">
										{stats?.totalRooms ? Math.round((stats.activeRooms / stats.totalRooms) * 100) : 0}%
									</span>
								</div>
								<div className="admin-progress-track">
									<span
										className="admin-progress-fill"
										style={{ width: `${stats?.totalRooms ? (stats.activeRooms / stats.totalRooms) * 100 : 0}%` }}
									/>
								</div>
							</div>
						</div>

						<div className="admin-card">
							<h3 className="admin-card-title">Recap Summary</h3>
							<p className="admin-card-subtitle">How the platform performed recently</p>
							<div className="admin-recap-grid">
								{[
									{ key: 'last24h', label: 'Last 24 Hours' },
									{ key: 'last7d', label: 'Last 7 Days' },
									{ key: 'last30d', label: 'Last 30 Days' },
								].map((period) => (
									<div className="admin-recap-card" key={period.key}>
										<span className="admin-recap-label">{period.label}</span>
										<div className="admin-recap-metric">
											<span>Users</span>
											<strong>{formatNumber(stats?.recaps?.[period.key]?.users)}</strong>
										</div>
										<div className="admin-recap-metric">
											<span>Rooms</span>
											<strong>{formatNumber(stats?.recaps?.[period.key]?.rooms)}</strong>
										</div>
										<div className="admin-recap-metric">
											<span>Transactions</span>
											<strong>{formatNumber(stats?.recaps?.[period.key]?.transactions)}</strong>
										</div>
										<div className="admin-recap-metric">
											<span>Revenue</span>
											<strong>{formatCurrency(stats?.recaps?.[period.key]?.revenue)}</strong>
										</div>
										<div className="admin-recap-metric">
											<span>Credits</span>
											<strong>{formatNumber(stats?.recaps?.[period.key]?.credits)}</strong>
										</div>
									</div>
								))}
							</div>
						</div>
					</div>

					<div className="admin-card">
						<h3 className="admin-card-title">Weekly Activity</h3>
						<p className="admin-card-subtitle">New users, rooms, and transactions in the last 7 days</p>
						<div className="admin-chart">
							<div className="admin-chart-header">
								<span>Users</span>
								<div className="admin-chart-labels">
									{stats?.series?.labels?.map((label) => (
										<span key={`user-label-${label}`}>{label}</span>
									))}
								</div>
							</div>
							{renderBars(stats?.series?.users, 'is-users')}
						</div>
						<div className="admin-chart">
							<div className="admin-chart-header">
								<span>Rooms</span>
								<div className="admin-chart-labels">
									{stats?.series?.labels?.map((label) => (
										<span key={`room-label-${label}`}>{label}</span>
									))}
								</div>
							</div>
							{renderBars(stats?.series?.rooms, 'is-rooms')}
						</div>
						<div className="admin-chart">
							<div className="admin-chart-header">
								<span>Transactions</span>
								<div className="admin-chart-labels">
									{stats?.series?.labels?.map((label) => (
										<span key={`txn-label-${label}`}>{label}</span>
									))}
								</div>
							</div>
							{renderBars(stats?.series?.transactions, 'is-transactions')}
						</div>
						<div className="admin-chart">
							<div className="admin-chart-header">
								<span>Revenue (USD)</span>
								<div className="admin-chart-labels">
									{stats?.series?.labels?.map((label) => (
										<span key={`rev-label-${label}`}>{label}</span>
									))}
								</div>
							</div>
							{renderBars(stats?.series?.revenue, 'is-revenue')}
						</div>
					</div>

					<div className="admin-card">
						<h3 className="admin-card-title">Platform Configuration</h3>
						<p className="admin-card-subtitle">Operational defaults and system metadata</p>
						<div className="admin-config-grid">
							<div className="admin-config-item">
								<span>Room Max Duration</span>
								<strong>{stats?.system?.roomMaxDuration} mins</strong>
							</div>
							<div className="admin-config-item">
								<span>Room Creation Cost</span>
								<strong>{stats?.system?.roomCreationCost} credits</strong>
							</div>
							<div className="admin-config-item">
								<span>Default Credits</span>
								<strong>{stats?.system?.defaultCredits} credits</strong>
							</div>
							<div className="admin-config-item">
								<span>Server Time</span>
								<strong>{stats?.serverTime}</strong>
							</div>
						</div>
					</div>
				</div>
			)}
		</AdminLayout>
	);
};

export default AdminDashboard;
