import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { checkRoomExists } from '../lib/roomStore.js';

const JoinRoom = () => {
	const [roomCode, setRoomCode] = useState('');
	const [error, setError] = useState('');
	const [isLoading, setIsLoading] = useState(false);
	const navigate = useNavigate();

	const handleSubmit = async (e) => {
		e.preventDefault();
		const code = roomCode.trim().toUpperCase();
		if (!code) {
			setError('Please enter a room code');
			return;
		}
		if (code.length < 4) {
			setError('Room code must be at least 4 characters');
			return;
		}

		setIsLoading(true);
		setError('');

		try {
			const result = await checkRoomExists(code);
			if (result.exists) {
				navigate(`/rooms/${code}/guest`);
			} else {
				setError('Room not found. Please check the code and try again.');
			}
		} catch {
			setError('Unable to verify room. Please try again.');
		} finally {
			setIsLoading(false);
		}
	};

	return (
		<div className="join-room-page">
			<div className="join-room-container">
				<div className="join-room-card">
					<Link to="/" className="join-room-back">
						← Back
					</Link>

					<div className="join-room-header">
						<div className="join-room-icon">♪</div>
						<h1>Join a Room</h1>
						<p className="join-room-subtitle">Enter the room code to join a karaoke session</p>
					</div>

					<form onSubmit={handleSubmit} className="join-room-form">
						<div className="join-room-field">
							<label htmlFor="roomCode">Room Code</label>
							<input
								id="roomCode"
								type="text"
								className="join-room-input"
								placeholder="e.g. ABC123"
								value={roomCode}
								onChange={(e) => {
									setRoomCode(e.target.value.toUpperCase());
									setError('');
								}}
								maxLength={10}
								autoFocus
							/>
							{error && <span className="join-room-error">{error}</span>}
						</div>

						<button type="submit" className="join-room-button" disabled={!roomCode.trim() || isLoading}>
							{isLoading ? <span className="auth-spinner" /> : 'Join Room'}
						</button>
					</form>

					<div className="join-room-footer">
						<p>Ask the host for the room code to join their session</p>
					</div>
				</div>

				<div className="join-room-ambient" />
			</div>
		</div>
	);
};

export default JoinRoom;
