import { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { fetchWithAuth, useAuth } from '../lib/auth.jsx';
import { createRoom, listRooms } from '../lib/roomStore.js';

const ADMIN_EMAIL = import.meta.env.VITE_ADMIN_EMAIL;
const API_BASE = (() => {
	const raw = import.meta.env.VITE_WS_HOST?.trim();
	if (raw && raw.length > 0) return raw.replace(/\/$/, '');
	if (typeof window !== 'undefined' && window.location) return window.location.origin.replace(/\/$/, '');
	return '';
})();

const Dashboard = () => {
	const { user, isAuthenticated, isLoading: authLoading, logout } = useAuth();
	const [rooms, setRooms] = useState([]);
	const [roomName, setRoomName] = useState('');
	const [isSubmitting, setIsSubmitting] = useState(false);
	const [createError, setCreateError] = useState(null);
	const [creditBalance, setCreditBalance] = useState(0);
	const [showExpiredRooms, setShowExpiredRooms] = useState(false);
	const [menuOpen, setMenuOpen] = useState(false);
	const navigate = useNavigate();

	const isAdmin = user?.email === ADMIN_EMAIL;

	useEffect(() => {
		let isActive = true;
		const fetchRooms = async () => {
			const items = await listRooms();
			if (isActive) {
				setRooms(items);
			}
		};
		if (isAuthenticated) {
			void fetchRooms();
		}
		return () => {
			isActive = false;
		};
	}, [isAuthenticated]);

	useEffect(() => {
		let isActive = true;
		const fetchCredits = async () => {
			if (!isAuthenticated) return;
			setCreditBalance(user?.credit || 0);
			try {
				const response = await fetchWithAuth(`${API_BASE}/api/credits`);
				if (!response.ok) return;
				const data = await response.json();
				if (isActive && typeof data.balance === 'number') {
					setCreditBalance(data.balance);
				}
			} catch {
				// ignore
			}
		};
		void fetchCredits();
		return () => {
			isActive = false;
		};
	}, [isAuthenticated, user?.credit]);

	const handleCreateRoom = async (event) => {
		event.preventDefault();
		if (!roomName.trim()) return;
		setIsSubmitting(true);
		setCreateError(null);
		try {
			const room = await createRoom(roomName.trim());
			const createdAt = room.createdAt || room.created_at || new Date().toISOString();
			const newRoom = {
				roomKey: room.roomKey,
				name: room.name || room.room_name || roomName.trim(),
				createdAt,
				expiredAt: room.expiredAt || room.expired_at || null,
			};
			setRooms((prevRooms) => {
				const filtered = prevRooms.filter((item) => item.roomKey !== newRoom.roomKey);
				return [newRoom, ...filtered];
			});
			void (async () => {
				try {
					const items = await listRooms();
					setRooms(items);
				} catch {
					// ignore
				}
			})();
			setRoomName('');
			navigate(`/rooms/${room.roomKey}`);
		} catch (err) {
			setCreateError(err.message || 'Failed to create room');
		} finally {
			setIsSubmitting(false);
		}
	};

	const handleLogout = async () => {
		await logout();
		navigate('/login');
	};

	const formatTimeRemaining = (room) => {
		if (!room.expiredAt) return null;
		const expiresAt = new Date(room.expiredAt);
		const diff = expiresAt - new Date();
		if (diff <= 0) return 'Expired';
		const minutes = Math.floor(diff / 60000);
		const hours = Math.floor(minutes / 60);
		if (hours > 0) return `${hours}h ${minutes % 60}m`;
		return `${minutes}m`;
	};

	const getRemainingMinutes = (room) => {
		if (!room.expiredAt) return 0;
		const expiresAt = new Date(room.expiredAt);
		const diff = expiresAt - new Date();
		if (diff <= 0) return 0;
		return Math.floor(diff / 60000);
	};

	const getTimeRemainingColor = (room) => {
		const minutes = getRemainingMinutes(room);
		if (minutes > 60) return 'green';
		if (minutes > 10) return 'yellow';
		return 'red';
	};

	const isRoomExpired = (room) => {
		if (!room.expiredAt) return false;
		const expiresAt = new Date(room.expiredAt);
		return expiresAt < new Date();
	};

	const expiredRooms = rooms.filter((room) => isRoomExpired(room));
	const activeRooms = rooms.filter((room) => !isRoomExpired(room));

	if (authLoading) {
		return (
			<div className="dashboard-page">
				<div className="dashboard-loading">
					<span className="auth-spinner" />
				</div>
			</div>
		);
	}

	return (
		<div className="dashboard-page">
			<header className="dashboard-header">
				<div className="dashboard-brand">
					<span className="dashboard-logo">♪</span>
					<span className="dashboard-title brand-logo-text">Karayouke</span>
				</div>
				<div className="dashboard-user">
					<button
						className="dashboard-menu-btn"
						onClick={() => setMenuOpen((prev) => !prev)}
						aria-label="Toggle menu"
						aria-expanded={menuOpen}
					>
						☰
					</button>
					<div className={`dashboard-user-items ${menuOpen ? 'open' : ''}`}>
					{isAuthenticated ? (
						<>
							{isAdmin && (
								<Link to="/admin" className="dashboard-admin-link">
									Admin
								</Link>
							)}
							<span className="dashboard-username">{user?.name}</span>
							<button className="dashboard-logout" onClick={handleLogout}>
								Sign out
							</button>
						</>
					) : (
						<Link to="/login" className="dashboard-login">
							Sign in
						</Link>
					)}
					</div>
				</div>
			</header>

			<main className="dashboard-main">
				<section className="dashboard-hero">
					<div className="dashboard-hero-content">
						<h1>
							{isAuthenticated ? `Welcome, ${user?.name?.split(' ')[0]}` : 'Collaborative Karaoke'}
						</h1>
						<p className="dashboard-hero-subtitle">
							{isAuthenticated
								? 'Create a room and let everyone queue the next hit.'
								: 'Sign in to create rooms and manage your karaoke sessions.'}
						</p>
						<Link to="/join" className="dashboard-join-btn">
							Join a Room
						</Link>
					</div>
				</section>

				{isAuthenticated ? (
					<>
						<section className="dashboard-section">
						<div className="dashboard-section-header">
							<h2>Account Info</h2>
						</div>
						<div className="dashboard-info-grid">
							<div className="dashboard-info-card">
								<div className="dashboard-info-label">Credits Available</div>
								<div className="dashboard-info-value">{creditBalance}</div>
								<div className="dashboard-info-subtext">Use credits to create rooms</div>
								<Link to="/packages" className="dashboard-buy-credits-btn">
									Buy Credits
								</Link>
							</div>
							<div className="dashboard-info-card">
								<div className="dashboard-info-label">Payment History</div>
								<div className="dashboard-info-value">📋</div>
								<div className="dashboard-info-subtext">View your transactions</div>
								<Link to="/payment/history" className="dashboard-history-btn">
									View History
								</Link>
							</div>
						</div>
					</section>

					<section className="dashboard-section">
						<div className="dashboard-card dashboard-create-card">
							<div className="dashboard-card-icon">+</div>
							<div className="dashboard-card-content">
								<h3>Create a new room</h3>
								<p>Start a karaoke session for your friends to join.</p>
							</div>
							<form onSubmit={handleCreateRoom} className="dashboard-create-form">
								<input
									id="room-name-input"
									type="text"
									className="dashboard-input"
									placeholder="Room name"
									value={roomName}
									onChange={(event) => setRoomName(event.target.value)}
								/>
								<button type="submit" className="dashboard-create-btn" disabled={isSubmitting}>
									{isSubmitting ? 'Creating…' : 'Create'}
								</button>
							</form>
							{createError && (
								<div className="dashboard-create-error">{createError}</div>
							)}
						</div>
					</section>

					<section className="dashboard-section">
						<div className="dashboard-section-header">
							<h2>Your rooms</h2>
							<div className="dashboard-rooms-actions">
								<span className="dashboard-badge">{activeRooms.length}</span>
								{expiredRooms.length > 0 && (
									<button
										className="dashboard-toggle"
										onClick={() => setShowExpiredRooms((prev) => !prev)}
									>
										{showExpiredRooms ? 'Hide expired' : 'Show expired'}
										<span className="dashboard-toggle-count">{expiredRooms.length}</span>
									</button>
								)}
							</div>
						</div>
						{activeRooms.length === 0 ? (
							<div className="dashboard-empty">
								<p>No rooms yet. Create one to get started.</p>
							</div>
						) : (
							<div className="dashboard-rooms-grid">
								{activeRooms.map((room) => {
								const timeLeft = formatTimeRemaining(room);
								const timeColor = getTimeRemainingColor(room);
								return (
									<Link
										to={`/rooms/${room.roomKey}`}
										key={room.roomKey}
										className="dashboard-room-card"
									>
										<div className="dashboard-room-info">
											<h4>{room.name}</h4>
											<span className="dashboard-room-date">
												{new Date(room.createdAt).toLocaleDateString()}
											</span>
											{timeLeft && timeLeft !== 'Expired' && (
												<span className={`dashboard-room-time ${timeColor}`}>{timeLeft} remaining</span>
											)}
											</div>
											<div className="dashboard-room-meta">
												<span className="dashboard-room-code">{room.roomKey}</span>
												<span className="dashboard-room-status active">Active</span>
											</div>
										</Link>
									);
								})}
							</div>
						)}
						{expiredRooms.length > 0 && showExpiredRooms && (
							<div className="dashboard-expired-block">
								<div className="dashboard-expired-header">Expired rooms</div>
								<div className="dashboard-rooms-grid">
									{expiredRooms.map((room) => (
										<div key={room.roomKey} className="dashboard-room-card expired">
											<div className="dashboard-room-info">
												<h4>{room.name}</h4>
												<span className="dashboard-room-date">
													{new Date(room.createdAt).toLocaleDateString()}
												</span>
												<span className="dashboard-room-time expired">Expired</span>
											</div>
											<div className="dashboard-room-meta">
												<span className="dashboard-room-code">{room.roomKey}</span>
												<span className="dashboard-room-status expired">Expired</span>
											</div>
										</div>
									))}
								</div>
							</div>
						)}
					</section>
					</>
				) : (
					<section className="dashboard-section dashboard-auth-prompt">
						<div className="dashboard-auth-card">
							<h3>Get started</h3>
							<p>Create an account or sign in to host your own karaoke rooms.</p>
							<div className="dashboard-auth-actions">
								<Link to="/register" className="dashboard-btn-primary">
									Create account
								</Link>
								<Link to="/login" className="dashboard-btn-secondary">
									Sign in
								</Link>
							</div>
						</div>
					</section>
				)}
			</main>

			<footer className="site-footer">
				<div className="site-footer-content">
					<div className="site-footer-brand">
						<span className="site-footer-logo">♪</span>
						<span className="site-footer-name">Karayouke</span>
					</div>
					<div className="site-footer-links">
						<Link to="/faq">FAQ</Link>
						<Link to="/terms">Terms &amp; Conditions</Link>
						<Link to="/refund-policy">Refund Policy</Link>
						<Link to="/credits">Credits</Link>
					</div>
					<div className="site-footer-contact">
						<p><a href="mailto:ask@karayouke.com">ask@karayouke.com</a></p>
						<p>Tamangapa Raya No. 43, Bangkala, Manggala, Kota Makassar, Sulawesi Selatan, Indonesia 90235</p>
					</div>
					<div className="site-footer-copy">
						<p>&copy; {new Date().getFullYear()} Karayouke. All rights reserved.</p>
					</div>
				</div>
			</footer>
		</div>
	);
};

export default Dashboard;
