import { useEffect, useMemo, useRef, useState } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { QRCodeCanvas } from 'qrcode.react';
import YouTube from 'react-youtube';
import { useRoom, checkRoomExists, subscribeToRoomExpiration } from '../lib/roomStore.js';

const RoomPlayer = () => {
	const { roomKey } = useParams();
	const { state, actions } = useRoom(roomKey);
	const navigate = useNavigate();
	const playerRef = useRef(null);
	const advanceGuardRef = useRef(false);
	const [isReady, setIsReady] = useState(false);
	const [roomError, setRoomError] = useState(null);
	const [roomInfo, setRoomInfo] = useState(null);
	const [isVerifying, setIsVerifying] = useState(true);
	const [isExpired, setIsExpired] = useState(false);

	const nowPlaying = state.nowPlaying || state.queue[0];
	const roomTitle = useMemo(() => roomInfo?.name || state.meta?.name || 'Player', [roomInfo?.name, state.meta?.name]);
	const guestUrl = useMemo(() => `${window.location.origin}/rooms/${roomKey}/guest`, [roomKey]);

	// Subscribe to room expiration
	useEffect(() => {
		const unsubscribe = subscribeToRoomExpiration(roomKey, () => {
			setIsExpired(true);
			// Redirect after showing message
			setTimeout(() => {
				navigate('/', { replace: true });
			}, 5000);
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

	// YouTube player options - matching Next_Karayouke exactly
	const opts = {
		playerVars: {
			start: 0,
			autoplay: 1,
			playsinline: 1,
			origin: window.location.origin,
			rel: 0,
			controls: 1,
		},
	};

	const onPlayerReady = (event) => {
		playerRef.current = event.target;
		setIsReady(true);
		try {
			event.target.unMute?.();
			event.target.setVolume?.(100);
			event.target.playVideo?.();
		} catch {
			// ignore
		}
	};

	const onPlayerStateChange = (event) => {
		if (event.data === window.YT.PlayerState.ENDED) {
			// Guard against double-advance: both onStateChange(ENDED) and onEnd can fire.
			// Only advance once per song ending, using a ref guard with a cooldown.
			if (advanceGuardRef.current) return;
			advanceGuardRef.current = true;
			setTimeout(() => { advanceGuardRef.current = false; }, 2000);
			actions.advanceSong();
			return;
		}
		if (event.data === window.YT.PlayerState.CUED || event.data === window.YT.PlayerState.UNSTARTED) {
			try {
				event.target.unMute?.();
				event.target.setVolume?.(100);
				event.target.playVideo?.();
			} catch {
				// ignore
			}
		}
		if (event.data === window.YT.PlayerState.PLAYING) {
			try {
				event.target.unMute?.();
				event.target.setVolume?.(100);
			} catch {
				// ignore
			}
		}
	};

	const onPlayerEnd = () => {
		// Guard against double-advance (same guard as onStateChange ENDED)
		if (advanceGuardRef.current) return;
		advanceGuardRef.current = true;
		setTimeout(() => { advanceGuardRef.current = false; }, 2000);
		actions.advanceSong();
	};

	// Note: No setNowPlaying effect needed here.
	// The player simply plays whatever nowPlaying is (first unplayed song).
	// Sending reorder commands from the player causes race conditions
	// when multiple users are adding songs to the queue.

	// Cleanup YouTube player on unmount to prevent memory leaks (critical for TV browsers)
	useEffect(() => {
		return () => {
			if (playerRef.current) {
				try {
					playerRef.current.destroy?.();
				} catch {
					// ignore errors during cleanup
				}
				playerRef.current = null;
			}
		};
	}, []);

	if (isVerifying) {
		return (
			<div className="player-shell">
				<div className="player-empty">
					<div className="player-empty-card" style={{ textAlign: 'center' }}>
						<span className="auth-spinner" />
						<p className="text-muted" style={{ marginTop: '16px' }}>Verifying room...</p>
					</div>
				</div>
			</div>
		);
	}

	if (isExpired) {
		return (
			<div className="player-shell">
				<div className="player-empty">
					<div className="player-empty-card" style={{ textAlign: 'center' }}>
						<h2 style={{ color: '#f97316', marginBottom: '16px' }}>⏰ Room Expired</h2>
						<p className="text-muted">This karaoke session has ended.</p>
						<p className="text-muted" style={{ fontSize: '0.9rem', marginTop: '12px' }}>Thanks for singing with us!</p>
					</div>
				</div>
			</div>
		);
	}
	if (roomError) {
		return (
			<div className="player-shell">
				<div className="player-empty">
					<div className="player-empty-card" style={{ textAlign: 'center' }}>
						<h2>Room Not Found</h2>
						<p className="text-muted">The room "{roomKey}" does not exist.</p>
						<Link to="/" style={{ display: 'inline-block', marginTop: '16px', padding: '12px 24px', background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)', borderRadius: '12px', color: '#fff', textDecoration: 'none', fontWeight: '600' }}>Back to Home</Link>
					</div>
				</div>
			</div>
		);
	}

	return (
		<div className="player-shell">
			{nowPlaying ? (
				<YouTube
					key={nowPlaying.id}
					videoId={nowPlaying.id}
					opts={opts}
					onReady={onPlayerReady}
					onStateChange={onPlayerStateChange}
					onEnd={onPlayerEnd}
					className="player-frame"
					iframeClassName="player-iframe"
				/>
			) : (
				<div className="player-empty">
					<div className="player-empty-card">
						<p className="eyebrow">{roomTitle}</p>
						<h2>Queue is empty</h2>
						<p className="text-muted">Scan to add songs as a guest.</p>
						<div className="qr-block">
							<QRCodeCanvas value={guestUrl} size={180} fgColor="#111" bgColor="#fff" includeMargin />
							<div className="stack">
								<p className="text-muted">{guestUrl}</p>
							</div>
						</div>
					</div>
				</div>
			)}
		</div>
	);
};

export default RoomPlayer;
