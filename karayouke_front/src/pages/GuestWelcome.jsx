import { useState, useEffect } from 'react';
import { useNavigate, useParams, Link } from 'react-router-dom';
import { useRoom, setGuestProfile, checkRoomExists, subscribeToRoomExpiration } from '../lib/roomStore.js';
import { useAuth } from '../lib/auth.jsx';

const GuestWelcome = () => {
	const { roomKey } = useParams();
	const { state } = useRoom(roomKey);
	const { user, isAuthenticated } = useAuth();
	const navigate = useNavigate();
	// Initialize guestName from user if authenticated
	const [guestName, setGuestName] = useState(() => (isAuthenticated && user?.name) ? user.name : '');
	const [isSubmitting, setIsSubmitting] = useState(false);
	const [roomInfo, setRoomInfo] = useState(null);
	const [isLoading, setIsLoading] = useState(true);
	const [roomError, setRoomError] = useState(null);
	const [isExpired, setIsExpired] = useState(false);

	// Subscribe to room expiration
	useEffect(() => {
		const unsubscribe = subscribeToRoomExpiration(roomKey, () => {
			setIsExpired(true);
			// Redirect guest to homepage after showing message
			setTimeout(() => {
				navigate('/', { replace: true });
			}, 3000);
		});
		return unsubscribe;
	}, [roomKey, navigate]);

	useEffect(() => {
		const verifyRoom = async () => {
			setIsLoading(true);
			try {
				const result = await checkRoomExists(roomKey);
				if (result.exists) {
					setRoomInfo(result);
					setRoomError(null);
					if (result.isExpired) {
						setIsExpired(true);
					}
				} else {
					setRoomError('Room not found');
				}
			} catch {
				setRoomError('Unable to verify room');
			} finally {
				setIsLoading(false);
			}
		};
		verifyRoom();
	}, [roomKey]);

	const handleJoin = async (e) => {
		e.preventDefault();
		if (!guestName.trim()) return;

		setIsSubmitting(true);
		await setGuestProfile(roomKey, { name: guestName.trim(), userId: user?.id || null });
		navigate(`/rooms/${roomKey}/guest/controller`);
	};

	if (isLoading) {
		return (
			<div className="guest-welcome-page">
				<div className="guest-welcome-container">
					<div className="guest-welcome-card" style={{ textAlign: 'center' }}>
						<span className="auth-spinner" />
						<p style={{ marginTop: '16px', color: 'rgba(148, 163, 184, 0.8)' }}>Verifying room...</p>
					</div>
					<div className="guest-welcome-ambient" />
				</div>
			</div>
		);
	}

	if (isExpired) {
		return (
			<div className="guest-welcome-page">
				<div className="guest-welcome-container">
					<div className="guest-welcome-card">
						<div className="guest-welcome-header">
							<div className="guest-welcome-icon" style={{ background: 'linear-gradient(135deg, #f97316 0%, #ea580c 100%)' }}>⏰</div>
							<h1>Room Expired</h1>
							<p className="guest-welcome-subtitle">This karaoke session has ended.</p>
						</div>
						<p style={{ textAlign: 'center', color: 'rgba(148, 163, 184, 0.6)', fontSize: '0.9rem' }}>Redirecting to homepage...</p>
					</div>
					<div className="guest-welcome-ambient" />
				</div>
			</div>
		);
	}

	if (roomError) {
		return (
			<div className="guest-welcome-page">
				<div className="guest-welcome-container">
					<div className="guest-welcome-card">
						<div className="guest-welcome-header">
							<div className="guest-welcome-icon" style={{ background: 'linear-gradient(135deg, #ef4444 0%, #dc2626 100%)' }}>✕</div>
							<h1>Room Not Found</h1>
							<p className="guest-welcome-subtitle">The room code "{roomKey}" does not exist or has been deleted.</p>
						</div>
						<Link to="/join" className="guest-welcome-button" style={{ textDecoration: 'none', textAlign: 'center' }}>
							Try Another Code
						</Link>
						<div className="guest-welcome-footer">
							<Link to="/" style={{ color: 'rgba(148, 163, 184, 0.8)', fontSize: '0.9rem' }}>← Back to Home</Link>
						</div>
					</div>
					<div className="guest-welcome-ambient" />
				</div>
			</div>
		);
	}

	return (
		<div className="guest-welcome-page">
			<div className="guest-welcome-container">
				<div className="guest-welcome-card">
					<div className="guest-welcome-header">
						<div className="guest-welcome-icon">♪</div>
						<h1>{roomInfo?.name || state.meta?.name || 'Karaoke Room'}</h1>
						<p className="guest-welcome-subtitle">Join and queue your favorite songs</p>
					</div>

					<form onSubmit={handleJoin} className="guest-welcome-form">
						<div className="guest-welcome-field">
							<label htmlFor="guestName">Your name</label>
							<input
								id="guestName"
								type="text"
								className="guest-welcome-input"
								placeholder="Enter your name"
								value={guestName}
								onChange={(e) => setGuestName(e.target.value)}
								required
								disabled={isAuthenticated}
							/>
							{isAuthenticated && (
								<span className="guest-welcome-hint">Signed in as {user?.name}</span>
							)}
						</div>

						<button type="submit" className="guest-welcome-button" disabled={isSubmitting || !guestName.trim()}>
							{isSubmitting ? (
								<span className="auth-spinner" />
							) : (
								'Enter Room'
							)}
						</button>
					</form>

					<div className="guest-welcome-footer">
						<p className="guest-welcome-code">Room code: <strong>{roomKey}</strong></p>
					</div>
				</div>

				<div className="guest-welcome-ambient" />
			</div>
		</div>
	);
};

export default GuestWelcome;
