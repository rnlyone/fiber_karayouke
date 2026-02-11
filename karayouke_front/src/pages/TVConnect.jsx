import { useEffect, useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useAuth, fetchWithAuth } from '../lib/auth.jsx';

const API_BASE = (() => {
	const raw = import.meta.env.VITE_WS_HOST?.trim();
	if (raw && raw.length > 0) return raw.replace(/\/$/, '');
	if (typeof window !== 'undefined' && window.location) return window.location.origin.replace(/\/$/, '');
	return '';
})();

// This page handles the QR code scan from a mobile device
// It allows the room master to select which room to connect the TV to
const TVConnect = () => {
	const { token } = useParams();
	const navigate = useNavigate();
	const { isAuthenticated, isLoading: authLoading } = useAuth();
	const [rooms, setRooms] = useState([]);
	const [loading, setLoading] = useState(true);
	const [connecting, setConnecting] = useState(false);
	const [error, setError] = useState(null);
	const [success, setSuccess] = useState(null);

	// Fetch user's rooms
	useEffect(() => {
		const fetchRooms = async () => {
			if (!isAuthenticated) return;
			try {
				const response = await fetchWithAuth(`${API_BASE}/api/rooms`);
				if (response.ok) {
					const data = await response.json();
					// Filter to only show active (non-expired) rooms
					const activeRooms = data.filter(room => {
						if (!room.expired_at) return true;
						return new Date(room.expired_at) > new Date();
					});
					setRooms(activeRooms);
				}
			} catch {
				setError('Failed to load rooms');
			} finally {
				setLoading(false);
			}
		};
		if (isAuthenticated) {
			fetchRooms();
		}
	}, [isAuthenticated]);

	// Redirect to login if not authenticated
	useEffect(() => {
		if (!authLoading && !isAuthenticated) {
			navigate('/login', { state: { from: `/tv/connect/${token}` } });
		}
	}, [authLoading, isAuthenticated, navigate, token]);

	const connectTV = async (roomKey) => {
		setConnecting(true);
		setError(null);
		try {
			const response = await fetchWithAuth(`${API_BASE}/api/tv/connect`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					code: token,
					room_key: roomKey,
				}),
			});
			const data = await response.json();
			if (!response.ok) {
				throw new Error(data.error || 'Failed to connect TV');
			}
			setSuccess(`TV connected to "${data.room_name}"!`);
			// Redirect to room master after success
			setTimeout(() => {
				navigate(`/rooms/${roomKey}`);
			}, 2000);
		} catch (err) {
			setError(err.message);
		} finally {
			setConnecting(false);
		}
	};

	if (authLoading || loading) {
		return (
			<div className="tv-connect-page">
				<div className="tv-connect-loading">
					<span className="auth-spinner" />
					<p>Loading...</p>
				</div>
			</div>
		);
	}

	if (success) {
		return (
			<div className="tv-connect-page">
				<div className="tv-connect-success">
					<div className="tv-connect-success-icon">✅</div>
					<h2>{success}</h2>
					<p>Redirecting to room...</p>
				</div>
			</div>
		);
	}

	return (
		<div className="tv-connect-page">
			<div className="tv-connect-content">
				<div className="tv-connect-header">
					<h1>📺 Connect TV</h1>
					<p>Select a room to display on the TV</p>
				</div>

				{error && (
					<div className="tv-connect-error">
						<p>{error}</p>
					</div>
				)}

				{rooms.length === 0 ? (
					<div className="tv-connect-empty">
						<p>You don't have any active rooms.</p>
						<Link to="/" className="tv-connect-btn-primary">
							Create a Room
						</Link>
					</div>
				) : (
					<div className="tv-connect-rooms">
						{rooms.map((room) => (
							<button
								key={room.room_key}
								className="tv-connect-room-btn"
								onClick={() => connectTV(room.room_key)}
								disabled={connecting}
							>
								<div className="tv-connect-room-info">
									<h3>{room.room_name || room.room_key}</h3>
									<span className="tv-connect-room-code">{room.room_key}</span>
								</div>
								<span className="tv-connect-room-arrow">→</span>
							</button>
						))}
					</div>
				)}

				<div className="tv-connect-footer">
					<Link to="/" className="tv-connect-back">
						← Back to Dashboard
					</Link>
				</div>
			</div>
		</div>
	);
};

export default TVConnect;
