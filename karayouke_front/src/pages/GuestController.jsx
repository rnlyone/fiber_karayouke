import { useEffect, useMemo, useRef, useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useRoom, getGuestProfile, checkRoomExists, subscribeToRoomExpiration } from '../lib/roomStore.js';
import { searchYoutube } from '../lib/youtube.js';
import Swal from 'sweetalert2';

const GuestController = () => {
	const { roomKey } = useParams();
	const { state, actions } = useRoom(roomKey);
	const navigate = useNavigate();
	const [searchTerm, setSearchTerm] = useState('');
	const [results, setResults] = useState([]);
	const [isSearching, setIsSearching] = useState(false);
	const [showFilters, setShowFilters] = useState(false);
	const [filterKaraokeOnly, setFilterKaraokeOnly] = useState(true);
	const [filterByViews, setFilterByViews] = useState(false);
	const [guestName, setGuestName] = useState('Guest');
	const [dragIndex, setDragIndex] = useState(null);
	const [dragTarget, setDragTarget] = useState(null);
	const [roomError, setRoomError] = useState(null);
	const [roomInfo, setRoomInfo] = useState(null);
	const [isVerifying, setIsVerifying] = useState(true);
	const [isExpired, setIsExpired] = useState(false);
	const dragSongIdRef = useRef(null);
	const dragTargetRef = useRef(null);
	const dragDroppedRef = useRef(false);
	const dragTimerRef = useRef(null);
	const dragActiveRef = useRef(false);

	const roomTitle = useMemo(() => roomInfo?.name || state.meta?.name || 'Controller', [roomInfo?.name, state.meta?.name]);

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

	// Verify room and load guest profile
	useEffect(() => {
		const loadProfile = async () => {
			setIsVerifying(true);

			// First verify room exists
			const roomResult = await checkRoomExists(roomKey);
			if (!roomResult.exists) {
				setRoomError('Room not found');
				setIsVerifying(false);
				return;
			}
			if (roomResult.isExpired) {
				setIsExpired(true);
				setIsVerifying(false);
				return;
			}
			setRoomInfo(roomResult);

			const profile = await getGuestProfile(roomKey);
			if (profile?.name) {
				setGuestName(profile.name);
			} else {
				// Redirect to welcome page if no profile
				navigate(`/rooms/${roomKey}/guest`);
			}
			setIsVerifying(false);
		};
		loadProfile();
	}, [roomKey, navigate]);

	if (isVerifying) {
		return (
			<div className="controller-page">
				<div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', minHeight: '100vh' }}>
					<span className="auth-spinner" />
				</div>
			</div>
		);
	}

	if (isExpired) {
		return (
			<div className="controller-page">
				<div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', minHeight: '100vh', gap: '16px' }}>
					<h2 style={{ color: '#f97316', fontSize: '1.5rem' }}>⏰ Room Expired</h2>
					<p style={{ color: 'rgba(148, 163, 184, 0.8)' }}>This karaoke session has ended.</p>
					<p style={{ color: 'rgba(148, 163, 184, 0.6)', fontSize: '0.9rem' }}>Redirecting to homepage...</p>
				</div>
			</div>
		);
	}

	if (roomError) {
		return (
			<div className="controller-page">
				<div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', minHeight: '100vh', gap: '16px' }}>
					<h2 style={{ color: '#f8fafc', fontSize: '1.5rem' }}>Room Not Found</h2>
					<p style={{ color: 'rgba(148, 163, 184, 0.8)' }}>The room "{roomKey}" does not exist.</p>
					<Link to="/" style={{ padding: '12px 24px', background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)', borderRadius: '12px', color: '#fff', textDecoration: 'none', fontWeight: '600' }}>Back to Home</Link>
				</div>
			</div>
		);
	}

	const handleSearch = async (event) => {
		event.preventDefault();
		if (!searchTerm.trim()) return;
		setIsSearching(true);
		const items = await searchYoutube(searchTerm.trim(), {
			includeKaraoke: filterKaraokeOnly,
			orderByViews: filterByViews,
		});
		setResults(items);
		setIsSearching(false);
	};

	const handleAddSong = async (song, position) => {
		actions.addSong(song, position);
		await Swal.fire({
			icon: 'success',
			title: 'Added to playlist!',
			showConfirmButton: false,
			timer: 1000,
			toast: true,
			position: 'top-end'
		});
	};

	const handleSkipSong = async () => {
		const result = await Swal.fire({
			icon: 'warning',
			title: 'Skip this song?',
			text: 'Are you sure you want to skip to the next song?',
			showCancelButton: true,
			confirmButtonColor: '#6366f1',
			cancelButtonColor: '#64748b',
			confirmButtonText: 'Yes, skip it',
			cancelButtonText: 'Cancel'
		});
		if (result.isConfirmed) {
			actions.skipSong();
		}
	};

	const handleExit = async (event) => {
		event.preventDefault();
		const result = await Swal.fire({
			icon: 'warning',
			title: 'Leave controller?',
			text: 'Are you sure you want to exit the controller?',
			showCancelButton: true,
			confirmButtonColor: '#6366f1',
			cancelButtonColor: '#64748b',
			confirmButtonText: 'Yes, leave',
			cancelButtonText: 'Stay'
		});
		if (result.isConfirmed) {
			navigate('/');
		}
	};

	const nowPlaying = state.nowPlaying || state.queue[0] || null;
	const upcoming = state.queue.filter((song) => song.id !== nowPlaying?.id);

	const resetDragState = () => {
		setDragIndex(null);
		setDragTarget(null);
		dragSongIdRef.current = null;
		dragTargetRef.current = null;
		dragDroppedRef.current = false;
		dragActiveRef.current = false;
		if (dragTimerRef.current) {
			clearTimeout(dragTimerRef.current);
			dragTimerRef.current = null;
		}
	};

	const getQueueIndexFromPoint = (clientX, clientY) => {
		const target = document.elementFromPoint(clientX, clientY);
		const item = target?.closest('[data-queue-index]');
		if (!item) return null;
		const nextTarget = Number(item.dataset.queueIndex);
		return Number.isNaN(nextTarget) ? null : nextTarget;
	};

	const startDrag = (songId, index, event) => {
		setDragIndex(index);
		setDragTarget(index);
		dragSongIdRef.current = songId;
		dragTargetRef.current = index;
		dragDroppedRef.current = false;
		dragActiveRef.current = true;
		if (event?.dataTransfer) {
			event.dataTransfer.effectAllowed = 'move';
			event.dataTransfer.setData('text/plain', songId);
			event.dataTransfer.setData('text/queue-index', String(index));
		}
	};

	const completeDrag = (targetIndex = null, sourceOverride = null) => {
		const resolvedTarget = targetIndex ?? dragTargetRef.current ?? dragTarget;
		const sourceSongId = sourceOverride ?? dragSongIdRef.current;

		if (sourceSongId === null || resolvedTarget === null) {
			resetDragState();
			return;
		}

		const sourceUpcomingIndex = upcoming.findIndex((song) => song.id === sourceSongId);
		if (sourceUpcomingIndex === -1 || sourceUpcomingIndex === resolvedTarget) {
			resetDragState();
			return;
		}

		const nowPlayingIndex = state.queue.findIndex((song) => song.id === nowPlaying?.id);
		const targetQueueIndex = nowPlayingIndex >= 0 ? resolvedTarget + 1 : resolvedTarget;
		actions.moveSong(sourceSongId, targetQueueIndex);
		resetDragState();
	};

	// Helper: check if mobile (max-width: 768px)
	const isMobile = typeof window !== 'undefined' && window.matchMedia && window.matchMedia('(max-width: 768px)').matches;

	// Move song up in queue
	const moveSongUp = (index) => {
		if (index <= 0) return;
		const songId = upcoming[index].id;
		const nowPlayingIndex = state.queue.findIndex((song) => song.id === nowPlaying?.id);
		const targetQueueIndex = nowPlayingIndex >= 0 ? index : index - 1;
		actions.moveSong(songId, targetQueueIndex);
	};

	// Move song down in queue
	const moveSongDown = (index) => {
		if (index >= upcoming.length - 1) return;
		const songId = upcoming[index].id;
		const nowPlayingIndex = state.queue.findIndex((song) => song.id === nowPlaying?.id);
		const targetQueueIndex = nowPlayingIndex >= 0 ? index + 2 : index + 1;
		actions.moveSong(songId, targetQueueIndex);
	};

	return (
		<div className="controller-shell controller-page">
			<div className="controller-top-grid">
				<div className="controller-info-card">
					<div className="controller-info-header">
						<Link to="/" onClick={handleExit} className="controller-info-icon" style={{ textDecoration: 'none', cursor: 'pointer' }} aria-label="Exit to home">
							←
						</Link>
						<div className="controller-info-details">
							<p className="controller-info-label">Room</p>
							<p className="controller-room-name">{roomTitle}</p>
						</div>
					</div>
					<div className="controller-info-meta">
						<div className="controller-meta-item">
							<span className="controller-meta-icon">♪</span>
							<span className="controller-meta-text">{guestName}</span>
						</div>
						<div className="controller-meta-item">
							<span className="controller-meta-icon">#</span>
							<span className="controller-meta-text">{roomKey}</span>
						</div>
					</div>
				</div>
				<div className="controller-now-playing-card">
					<p className="controller-label">Now playing</p>
					<h1>{nowPlaying ? nowPlaying.title : 'Playlist is empty'}</h1>
					<p className="text-muted">{nowPlaying?.artist || 'No artist information'}</p>
				</div>
			</div>

			<div className="controller-grid">
				<section className="controller-card controller-search-card">
					<h2>Search &amp; Queue</h2>
					<form onSubmit={handleSearch} className="search-form">
						<input
							type="text"
							className="input search-input"
							placeholder="Search for a karaoke track"
							value={searchTerm}
							onChange={(event) => setSearchTerm(event.target.value)}
						/>
						<button className="btn btn-primary search-btn" type="submit" disabled={isSearching} aria-label="Search">
							{isSearching ? (
								<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
									<circle cx="12" cy="12" r="10" opacity="0.3"/>
									<path d="M12 6v6l4 2"/>
								</svg>
							) : (
								<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
									<circle cx="11" cy="11" r="8"/>
									<path d="m21 21-4.35-4.35"/>
								</svg>
							)}
						</button>
					</form>

				<button 
					type="button" 
					className="filter-btn" 
					onClick={() => setShowFilters(!showFilters)}
					aria-label="Toggle filters"
				>
					<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
						<polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3"/>
					</svg>
					Filters
				</button>

				{showFilters && (
					<div className="filter-modal">
						<div className="filter-overlay" onClick={() => setShowFilters(false)} />
						<div className="filter-content">
							<h3>Search Filters</h3>
							<label className="filter-checkbox-label">
								<input 
									type="checkbox" 
									checked={filterKaraokeOnly}
									onChange={(e) => setFilterKaraokeOnly(e.target.checked)}
								/>
								<span>Display karaoke only videos</span>
							</label>
							<label className="filter-checkbox-label">
								<input 
									type="checkbox" 
									checked={filterByViews}
									onChange={(e) => setFilterByViews(e.target.checked)}
								/>
								<span>Order from most viewed video</span>
							</label>
							<button className="filter-close-btn" onClick={() => setShowFilters(false)}>Done</button>
						</div>
					</div>
				)}

				<div className="controller-list controller-results-grid scroll-fade">
					{results.map((song) => (
							<div key={song.id} className="video-card">
								<div className="video-card-thumbnail">
									<img src={song.thumbnail || 'https://placehold.co/320x180?text=Karaoke'} alt={song.title} />
									<div className="video-card-gradient" />
									<span className="video-duration">{song.duration || '—'}</span>
									<div className="video-card-actions">
										<button className="btn-card-action" onClick={() => handleAddSong(song, 'next')}>
											Play Next
										</button>
										<button className="btn-card-add" onClick={() => handleAddSong(song)}>
											+
										</button>
									</div>
									<div className="video-card-info">
										<p className="video-card-title">{song.title}</p>
										<p className="video-card-artist">{song.artist || 'YouTube'}</p>
									</div>
								</div>
							</div>
						))}
						{results.length === 0 && !isSearching && (
							<p className="text-muted">Search to see karaoke results.</p>
						)}
					</div>
				</section>

				<aside className="controller-card controller-upnext-card">
					<div className="controller-upnext">
						<span className="controller-upnext-icon" aria-hidden="true">♪</span>
						<div>
							<p className="controller-label">Up next</p>
							<p className="controller-upnext-title">{upcoming[0]?.title ?? 'No upcoming songs'}</p>
						</div>
					</div>
					<div className="controller-queue scroll-fade">
						{upcoming.length === 0 ? (
							<p className="text-muted">No additional songs queued.</p>
						) : (
							<ul
								className="controller-results-grid"
								onDragOver={(event) => {
									event.preventDefault();
									const nextTarget = getQueueIndexFromPoint(event.clientX, event.clientY);
									if (nextTarget === null) return;
									setDragTarget(nextTarget);
									dragTargetRef.current = nextTarget;
								}}
								onDrop={(event) => {
									event.preventDefault();
									dragDroppedRef.current = true;
									const nextTarget = getQueueIndexFromPoint(event.clientX, event.clientY);
									const sourceSongId = event.dataTransfer?.getData('text/plain') || dragSongIdRef.current;
									completeDrag(nextTarget ?? dragTargetRef.current, sourceSongId || null);
								}}
							>
								{upcoming.map((song, index) => {
									const isDragging = dragIndex === index;
									return (
										<li
											key={song.id}
											className={`queue-item ${isDragging ? 'queue-item-dragging' : ''}`}
											data-queue-index={index}
											draggable
											onDragStart={(event) => startDrag(song.id, index, event)}
											onDragOver={(event) => {
												event.preventDefault();
												setDragTarget(index);
												dragTargetRef.current = index;
											}}
											onDragEnter={(event) => {
												event.preventDefault();
												setDragTarget(index);
												dragTargetRef.current = index;
											}}
											onDrop={(event) => {
												event.preventDefault();
												dragDroppedRef.current = true;
												const sourceSongId = event.dataTransfer?.getData('text/plain') || dragSongIdRef.current;
												completeDrag(index, sourceSongId || null);
											}}
											onDragEnd={() => {
												if (!dragDroppedRef.current) {
													completeDrag(dragTargetRef.current);
													return;
												}
												resetDragState();
											}}
											onPointerDown={(event) => {
												if (event.pointerType !== 'touch') return;
												if (event.target.closest('.queue-actions')) return;
												event.preventDefault();
												dragTimerRef.current = setTimeout(() => {
													startDrag(song.id, index);
												}, 200);
											}}
											onPointerMove={(event) => {
												if (!dragActiveRef.current) return;
												const target = document.elementFromPoint(event.clientX, event.clientY);
												const item = target?.closest('[data-queue-index]');
												if (!item) return;
												const nextTarget = Number(item.dataset.queueIndex);
												setDragTarget(nextTarget);
												dragTargetRef.current = nextTarget;
											}}
											onPointerUp={() => completeDrag()}
											onPointerCancel={() => resetDragState()}
										>
											<div className="queue-meta">
												<span className="queue-index">#{index + 2}</span>
												<div className="queue-info">
													<span className="queue-title">{song.title}</span>
													{song.singerName && (
														<span className="queue-singer">♪ : {song.singerName}</span>
													)}
												</div>
											</div>
											<div className="queue-actions">
												{isMobile && (
													<>
														<button
															type="button"
															className="queue-move queue-move-up"
															aria-label="Move up"
															disabled={index === 0}
															onClick={() => moveSongUp(index)}
														>
															▲
														</button>
														<button
															type="button"
															className="queue-move queue-move-down"
															aria-label="Move down"
															disabled={index === upcoming.length - 1}
															onClick={() => moveSongDown(index)}
														>
															▼
														</button>
													</>
												)}
												<button
													type="button"
													className="queue-remove"
													aria-label="Remove song"
													onClick={() => actions.removeSong(song.id)}
												>
													×
												</button>
											</div>
										</li>
									);
								})}
							</ul>
						)}
					</div>
					<button className="btn btn-primary" onClick={handleSkipSong}>
						Skip to next
					</button>
				</aside>
			</div>
		</div>
	);
};

export default GuestController;
