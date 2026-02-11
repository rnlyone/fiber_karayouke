import { useEffect, useMemo, useState, useRef, useCallback } from 'react';
import { Link, useParams, useNavigate } from 'react-router-dom';
import { QRCodeCanvas } from 'qrcode.react';
import { useRoom, checkRoomExists, subscribeToRoomExpiration } from '../lib/roomStore.js';
import { useAuth, fetchWithAuth } from '../lib/auth.jsx';

const API_BASE = (() => {
	const raw = import.meta.env.VITE_WS_HOST?.trim();
	if (raw && raw.length > 0) return raw.replace(/\/$/, '');
	if (typeof window !== 'undefined' && window.location) return window.location.origin.replace(/\/$/, '');
	return '';
})();

const RoomMaster = () => {
	const { roomKey } = useParams();
	const { state } = useRoom(roomKey);
	const { isAuthenticated, isLoading } = useAuth();
	const navigate = useNavigate();
	const [roomInfo, setRoomInfo] = useState(null);
	const [roomError, setRoomError] = useState(null);
	const [isVerifying, setIsVerifying] = useState(true);
	const [isExpired, setIsExpired] = useState(false);
	
	// TV Connection state
	const [tvCode, setTvCode] = useState('');
	const [tvConnecting, setTvConnecting] = useState(false);
	const [tvError, setTvError] = useState(null);
	const [tvSuccess, setTvSuccess] = useState(null);
	const [showScanner, setShowScanner] = useState(false);
	const videoRef = useRef(null);
	const streamRef = useRef(null);

	const guestUrl = useMemo(() => `${window.location.origin}/rooms/${roomKey}/guest`, [roomKey]);
	const controllerUrl = useMemo(() => `/rooms/${roomKey}/controller`, [roomKey]);
	const playerUrl = useMemo(() => `/rooms/${roomKey}/player`, [roomKey]);

	// TV Connection handlers (defined before hooks that use them)
	const stopScanner = useCallback(() => {
		if (streamRef.current) {
			streamRef.current.getTracks().forEach(track => track.stop());
			streamRef.current = null;
		}
		setShowScanner(false);
	}, []);

	const connectTV = useCallback(async (code) => {
		if (!code || code.trim().length === 0) {
			setTvError('Please enter a TV code');
			return;
		}
		setTvConnecting(true);
		setTvError(null);
		setTvSuccess(null);
		try {
			const response = await fetchWithAuth(`${API_BASE}/api/tv/connect`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					code: code.trim(),
					room_key: roomKey,
				}),
			});
			const data = await response.json();
			if (!response.ok) {
				throw new Error(data.error || 'Failed to connect TV');
			}
			setTvSuccess('TV connected successfully!');
			setTvCode('');
			setTimeout(() => setTvSuccess(null), 3000);
		} catch (err) {
			setTvError(err.message);
		} finally {
			setTvConnecting(false);
		}
	}, [roomKey]);

	const handleTvCodeSubmit = useCallback((e) => {
		e.preventDefault();
		connectTV(tvCode);
	}, [connectTV, tvCode]);

	const startScanner = useCallback(async () => {
		setShowScanner(true);
		setTvError(null);
		try {
			const stream = await navigator.mediaDevices.getUserMedia({ 
				video: { facingMode: 'environment' } 
			});
			streamRef.current = stream;
			if (videoRef.current) {
				videoRef.current.srcObject = stream;
			}
		} catch {
			setTvError('Camera access denied. Please allow camera access to scan QR codes.');
			setShowScanner(false);
		}
	}, []);

	// Subscribe to room expiration
	useEffect(() => {
		const unsubscribe = subscribeToRoomExpiration(roomKey, () => {
			setIsExpired(true);
			// Redirect master to dashboard after showing message
			setTimeout(() => {
				navigate('/', { replace: true });
			}, 3000);
		});
		return unsubscribe;
	}, [roomKey, navigate]);

	// Verify room exists
	useEffect(() => {
		const verifyRoom = async () => {
			setIsVerifying(true);
			try {
				const result = await checkRoomExists(roomKey);
				if (result.exists) {
					setRoomInfo(result);
					setRoomError(null);
					// Check if already expired
					if (result.isExpired) {
						setIsExpired(true);
					}
				} else {
					setRoomError('Room not found');
				}
			} catch {
				setRoomError('Unable to verify room');
			} finally {
				setIsVerifying(false);
			}
		};
		verifyRoom();
	}, [roomKey]);

	// Redirect non-authenticated users to guest page
	useEffect(() => {
		if (!isLoading && !isAuthenticated && !isVerifying && !roomError) {
			navigate(`/rooms/${roomKey}/guest`, { replace: true });
		}
	}, [isLoading, isAuthenticated, roomKey, navigate, isVerifying, roomError]);

	// Scan loop when scanner is active
	useEffect(() => {
		if (!showScanner) return;
		
		const scanQRCode = async () => {
			if (!videoRef.current) return;
			
			const video = videoRef.current;
			if (video.readyState !== video.HAVE_ENOUGH_DATA) return;

			const canvas = document.createElement('canvas');
			canvas.width = video.videoWidth;
			canvas.height = video.videoHeight;
			const ctx = canvas.getContext('2d');
			ctx.drawImage(video, 0, 0);

			// Use BarcodeDetector if available (modern browsers)
			if ('BarcodeDetector' in window) {
				try {
					const barcodeDetector = new window.BarcodeDetector({ formats: ['qr_code'] });
					const barcodes = await barcodeDetector.detect(canvas);
					if (barcodes.length > 0) {
						const url = barcodes[0].rawValue;
						// Extract token from URL: /tv/connect/{token}
						const match = url.match(/\/tv\/connect\/([a-f0-9]+)/i);
						if (match) {
							stopScanner();
							connectTV(match[1]);
							return;
						}
					}
				} catch {
					// BarcodeDetector failed, continue
				}
			}
		};

		const interval = setInterval(scanQRCode, 500);
		return () => clearInterval(interval);
	}, [showScanner, stopScanner, connectTV]);

	// Cleanup scanner on unmount
	useEffect(() => {
		return () => {
			if (streamRef.current) {
				streamRef.current.getTracks().forEach(track => track.stop());
			}
		};
	}, []);

	if (isLoading || isVerifying) {
		return (
			<div className="room-master-page">
				<div className="room-master-loading">
					<span className="auth-spinner" />
				</div>
			</div>
		);
	}

	if (isExpired) {
		return (
			<div className="room-master-page">
				<div className="room-master-loading" style={{ flexDirection: 'column', gap: '16px' }}>
					<h2 style={{ color: '#f87171', fontSize: '1.5rem' }}>⏰ Room Expired</h2>
					<p style={{ color: 'rgba(148, 163, 184, 0.8)' }}>This room session has ended.</p>
					<p style={{ color: 'rgba(148, 163, 184, 0.6)', fontSize: '0.9rem' }}>Redirecting to dashboard...</p>
				</div>
			</div>
		);
	}

	if (roomError) {
		return (
			<div className="room-master-page">
				<div className="room-master-loading" style={{ flexDirection: 'column', gap: '16px' }}>
					<h2 style={{ color: '#f8fafc', fontSize: '1.5rem' }}>Room Not Found</h2>
					<p style={{ color: 'rgba(148, 163, 184, 0.8)' }}>The room "{roomKey}" does not exist.</p>
					<Link to="/" className="room-master-btn-primary">Back to Dashboard</Link>
				</div>
			</div>
		);
	}

	if (!isAuthenticated) {
		return null;
	}

	return (
		<div className="room-master-page">
			<header className="room-master-header">
				<div className="room-master-brand">
					<Link to="/" className="room-master-back">
						<span>←</span>
					</Link>
					<div>
						<h1>{roomInfo?.name || state.meta?.name || 'Room'}</h1>
						<span className="room-master-code">{roomKey}</span>
					</div>
				</div>
			</header>

			<main className="room-master-main">
				<section className="room-master-hero">
					<div className="room-master-hero-content">
						<p className="room-master-eyebrow">Room master</p>
						<h2>Invite guests, manage the queue, and launch the player.</h2>
						<p className="room-master-subtitle">Share the guest link so everyone can search and add songs.</p>
					</div>
					<div className="room-master-actions">
						<Link to={controllerUrl} className="room-master-btn-primary">
							Open controller
						</Link>
						<Link to={playerUrl} className="room-master-btn-secondary">
							Open player
						</Link>
					</div>
				</section>

				<div className="room-master-grid">
					<section className="room-master-card">
						<div className="room-master-card-header">
							<h3>Guest access</h3>
							<p>Let guests scan the QR code to add songs.</p>
						</div>
						<div className="room-master-qr">
							<QRCodeCanvas value={guestUrl} size={180} fgColor="#111" bgColor="#fff" includeMargin />
							<div className="room-master-qr-info">
								<p className="room-master-qr-url">{guestUrl}</p>
								<a className="room-master-btn-secondary" href={guestUrl} target="_blank" rel="noreferrer">
									Open guest link
								</a>
							</div>
						</div>
					</section>

					<section className="room-master-card">
						<div className="room-master-card-header">
							<h3>Control links</h3>
							<p>Open the admin controller or player on a second screen.</p>
						</div>
						<div className="room-master-links">
							<Link className="room-master-btn-primary" to={controllerUrl}>
								Controller
							</Link>
							<Link className="room-master-btn-secondary" to={playerUrl}>
								Player
							</Link>
							<p className="room-master-hint">Keep the player visible for uninterrupted autoplay.</p>
						</div>
					</section>

					<section className="room-master-card room-master-tv-card">
						<div className="room-master-card-header">
							<h3>📺 Connect TV</h3>
							<p>Link a TV player to this room for the big screen experience.</p>
						</div>
						<div className="room-master-tv-connect">
							{tvSuccess && (
								<div className="room-master-tv-success">
									<span>✅</span> {tvSuccess}
								</div>
							)}
							{tvError && (
								<div className="room-master-tv-error">
									<span>⚠️</span> {tvError}
								</div>
							)}
							
							{showScanner ? (
								<div className="room-master-scanner">
									<video 
										ref={videoRef} 
										autoPlay 
										playsInline 
										muted
										className="room-master-scanner-video"
									/>
									<div className="room-master-scanner-overlay">
										<div className="room-master-scanner-frame" />
									</div>
									<button 
										className="room-master-btn-secondary" 
										onClick={stopScanner}
									>
										Cancel Scan
									</button>
								</div>
							) : (
								<>
									<form onSubmit={handleTvCodeSubmit} className="room-master-tv-form">
										<input
											type="text"
											value={tvCode}
											onChange={(e) => setTvCode(e.target.value.toUpperCase())}
											placeholder="Enter 5-char code"
											maxLength={5}
											className="room-master-tv-input"
											disabled={tvConnecting}
										/>
										<button 
											type="submit" 
											className="room-master-btn-primary"
											disabled={tvConnecting || tvCode.length !== 5}
										>
											{tvConnecting ? '...' : 'Connect'}
										</button>
									</form>
									<div className="room-master-tv-divider">
										<span>or</span>
									</div>
									<button 
										className="room-master-btn-secondary room-master-scan-btn"
										onClick={startScanner}
									>
										<span className="room-master-scan-icon">📷</span>
										Scan QR Code
									</button>
								</>
							)}
							<p className="room-master-hint">
								Open <a href="/tv" target="_blank" rel="noreferrer">karayouke.com/tv</a> on your TV browser to get the code.
							</p>
						</div>
					</section>
				</div>
			</main>
		</div>
	);
};

export default RoomMaster;
