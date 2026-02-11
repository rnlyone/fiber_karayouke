import { useEffect, useMemo, useState } from 'react';
import { getAuthToken } from './auth.jsx';

const ROOMS_KEY = 'karayouke:rooms';
const connections = new Map();

// Get API base URL from environment
const getApiBase = () => {
	const raw = import.meta.env.VITE_WS_HOST?.trim();
	if (raw && raw.length > 0) {
		return raw.replace(/\/$/, '');
	}
	if (typeof window !== 'undefined' && window.location) {
		return window.location.origin.replace(/\/$/, '');
	}
	return '';
};

const delay = (ms = 0) => new Promise((resolve) => setTimeout(resolve, ms));

const safeParse = (value, fallback) => {
	if (value === null || value === undefined || value === '') {
		return fallback;
	}
	try {
		return JSON.parse(value);
	} catch {
		return fallback;
	}
};

const _saveRooms = (rooms) => {
	localStorage.setItem(ROOMS_KEY, JSON.stringify(rooms));
};

// Get WebSocket host from environment or default to Go server
const getWebSocketHost = () => {
	const raw = import.meta.env.VITE_WS_HOST?.trim();
	if (raw && raw.length > 0) {
		try {
			const parsed = new URL(raw);
			if (parsed.protocol === 'https:') {
				parsed.protocol = 'wss:';
			} else if (parsed.protocol === 'http:') {
				parsed.protocol = 'ws:';
			}
			return `${parsed.protocol}//${parsed.host}`;
		} catch {
			const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
			const sanitized = raw.replace(/^(https?:|wss?:)\/\//, '');
			return `${protocol}//${sanitized}`;
		}
	}
	const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
	return `${protocol}//${window.location.host}`;
};

const ensureConnection = (roomKey) => {
	if (connections.has(roomKey)) {
		return connections.get(roomKey);
	}

	let state = null;
	let resolveReady;
	const ready = new Promise((resolve) => {
		resolveReady = resolve;
	});
	const listeners = new Set();
	const expiredListeners = new Set();
	const emojiListeners = new Set();

	// Use native WebSocket instead of PartySocket
	const wsUrl = `${getWebSocketHost()}/ws/${roomKey}`;
	const socket = new WebSocket(wsUrl);

	const notify = (nextState) => {
		state = nextState;
		listeners.forEach((callback) => callback(nextState));
		if (resolveReady) {
			resolveReady(nextState);
			resolveReady = null;
		}
	};

	const notifyExpired = () => {
		expiredListeners.forEach((callback) => callback());
	};

	const notifyEmoji = (emoji) => {
		emojiListeners.forEach((callback) => callback(emoji));
	};

	socket.addEventListener('open', () => {
		socket.send(JSON.stringify({ type: 'getState' }));
	});

	socket.addEventListener('message', (event) => {
		let payload;
		try {
			payload = JSON.parse(event.data);
		} catch {
			return;
		}
		if (payload?.type === 'state') {
			notify(payload.state || { playlist: [], meta: null });
		} else if (payload?.type === 'room_expired') {
			// Room has expired, notify listeners
			notifyExpired();
		} else if (payload?.type === 'error' && payload.error === 'Room not found') {
			// Room not found, treat as expired
			notifyExpired();
		} else if (payload?.type === 'emoji' && payload.emoji) {
			notifyEmoji(payload.emoji);
		}
	});

	// Handle reconnection
	socket.addEventListener('close', () => {
		connections.delete(roomKey);
		// Attempt to reconnect after a delay
		setTimeout(() => {
			if (listeners.size > 0) {
				const newConn = ensureConnection(roomKey);
				// Transfer listeners
				listeners.forEach((cb) => newConn.listeners.add(cb));
				expiredListeners.forEach((cb) => newConn.expiredListeners.add(cb));
				emojiListeners.forEach((cb) => newConn.emojiListeners.add(cb));
			}
		}, 2000);
	});

	const connection = { socket, ready, listeners, expiredListeners, emojiListeners, getState: () => state };
	connections.set(roomKey, connection);
	return connection;
};

const sendAction = async (roomKey, action) => {
	const connection = ensureConnection(roomKey);
	await connection.ready;
	if (connection.socket.readyState === WebSocket.OPEN) {
		connection.socket.send(JSON.stringify(action));
		return;
	}

	const sendOnOpen = () => {
		connection.socket.send(JSON.stringify(action));
		connection.socket.removeEventListener('open', sendOnOpen);
	};
	connection.socket.addEventListener('open', sendOnOpen);
};

export const listRooms = async () => {
	const token = getAuthToken();
	if (!token) {
		return [];
	}
	try {
		const response = await fetch(`${getApiBase()}/api/rooms`, {
			headers: { Authorization: `Bearer ${token}` },
		});
		if (!response.ok) {
			return [];
		}
		const rooms = await response.json();
		// Map API response to local format
		return rooms.map((room) => ({
			roomKey: room.room_key,
			name: room.room_name,
			createdAt: room.created_at,
			expiredAt: room.expired_at,
		}));
	} catch {
		return [];
	}
};

export const createRoom = async (name) => {
	const token = getAuthToken();
	if (!token) {
		throw new Error('Authentication required to create a room');
	}

	const response = await fetch(`${getApiBase()}/api/rooms`, {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
			Authorization: `Bearer ${token}`,
		},
		body: JSON.stringify({ name }),
	});

	if (!response.ok) {
		const error = await response.json();
		throw new Error(error.error || 'Failed to create room');
	}

	const room = await response.json();
	const roomKey = room.room_key;

	// Initialize WebSocket room state with metadata
	sendAction(roomKey, { type: 'setRoomMeta', name: room.room_name, createdAt: room.created_at });

	return {
		roomKey,
		name: room.room_name,
		createdAt: room.created_at,
		playlist: [],
		nowPlaying: null,
	};
};

// Check if a room exists in the database
export const checkRoomExists = async (roomKey) => {
	try {
		const response = await fetch(`${getApiBase()}/api/rooms/${roomKey}`);
		if (response.ok) {
			const room = await response.json();
			return {
				exists: true,
				roomKey: room.room_key,
				name: room.room_name,
				createdAt: room.created_at,
				expiredAt: room.expired_at,
				isExpired: room.is_expired || false,
			};
		}
		return { exists: false };
	} catch {
		return { exists: false };
	}
};

export const getRoomData = async (roomKey) => {
	const connection = ensureConnection(roomKey);
	if (connection.getState()) {
		return connection.getState();
	}
	return connection.ready;
};

export const subscribeToRoom = (roomKey, callback) => {
	const connection = ensureConnection(roomKey);
	const listener = (state) => callback(state);
	connection.listeners.add(listener);
	if (connection.getState()) {
		callback(connection.getState());
	}
	return () => connection.listeners.delete(listener);
};

export const subscribeToRoomExpiration = (roomKey, callback) => {
	const connection = ensureConnection(roomKey);
	connection.expiredListeners.add(callback);
	return () => connection.expiredListeners.delete(callback);
};

export const sendEmoji = async (roomKey, emoji) => {
	await sendAction(roomKey, { type: 'emoji', emoji });
};

export const subscribeToEmoji = (roomKey, callback) => {
	const connection = ensureConnection(roomKey);
	connection.emojiListeners.add(callback);
	return () => connection.emojiListeners.delete(callback);
};

export const updateRoom = async (roomKey, patch) => {
	await sendAction(roomKey, { type: 'updateRoom', patch });
	return null;
};

export const addSong = async (roomKey, song, profile, insertPosition) => {
	const resolvedProfile = profile || (await getGuestProfile(roomKey));
	await sendAction(roomKey, {
		type: 'add-video',
		id: song.id,
		title: song.title,
		artist: song.artist,
		coverUrl: song.thumbnail || 'https://placehold.co/320x320?text=Karaoke',
		singerName: resolvedProfile?.name || 'Guest',
		insertPosition,
	});
};

export const removeSong = async (roomKey, songId) => {
	await sendAction(roomKey, { type: 'remove-video', id: songId });
};

export const moveSong = async (roomKey, songId, targetIndex) => {
	const connection = ensureConnection(roomKey);
	const current = connection.getState();
	const playlist = current?.playlist || [];
	const unplayed = playlist.filter((v) => !v.playedAt);
	const idx = unplayed.findIndex((v) => v.id === songId);
	if (idx === -1) return;
	const clamped = Math.max(0, Math.min(unplayed.length - 1, targetIndex));
	const updated = [...unplayed];
	const [item] = updated.splice(idx, 1);
	updated.splice(clamped, 0, item);
	await sendAction(roomKey, { type: 'reorder-upcoming', ids: updated.map((v) => v.id) });
};

export const setNowPlaying = async (roomKey, song) => {
	const connection = ensureConnection(roomKey);
	const playlist = connection.getState()?.playlist || [];
	const unplayed = playlist.filter((v) => !v.playedAt);
	const idx = unplayed.findIndex((v) => v.id === song.id);
	if (idx === -1) return;
	const updated = [...unplayed];
	const [item] = updated.splice(idx, 1);
	updated.unshift(item);
	await sendAction(roomKey, { type: 'reorder-upcoming', ids: updated.map((v) => v.id) });
};

export const skipSong = async (roomKey) => {
	const connection = ensureConnection(roomKey);
	const playlist = connection.getState()?.playlist || [];
	const currentVideo = playlist.find((v) => !v.playedAt);
	if (!currentVideo) return;
	await sendAction(roomKey, { type: 'mark-as-played', id: currentVideo.id });
};

export const advanceSong = async (roomKey) => skipSong(roomKey);

export const setGuestProfile = async (roomKey, profile) => {
	await delay();
	localStorage.setItem(`karayouke:guest:${roomKey}`, JSON.stringify(profile));
};

export const getGuestProfile = async (roomKey) => {
	await delay();
	return safeParse(localStorage.getItem(`karayouke:guest:${roomKey}`), null);
};

export const useRoom = (roomKey) => {
	const [state, setState] = useState({ meta: null, queue: [], nowPlaying: null, playlist: [] });

	useEffect(() => {
		if (!roomKey) return undefined;
		const unsubscribe = subscribeToRoom(roomKey, (nextState) => {
			const playlist = nextState.playlist ?? [];
			const nowPlaying = playlist.find((v) => !v.playedAt) ?? null;
			const queue = playlist.filter((v) => !v.playedAt);
			setState({
				meta: nextState.meta ?? null,
				playlist,
				queue,
				nowPlaying,
			});
		});
		return () => unsubscribe();
	}, [roomKey]);

	const actions = useMemo(
		() => ({
			setRoomMeta: (name) => updateRoom(roomKey, { meta: { name } }),
			addSong: (song, insertPosition) => addSong(roomKey, song, null, insertPosition),
			removeSong: (songId) => removeSong(roomKey, songId),
			moveSong: (songId, direction) => moveSong(roomKey, songId, direction),
			setNowPlaying: (song) => setNowPlaying(roomKey, song),
			skipSong: () => skipSong(roomKey),
			advanceSong: () => advanceSong(roomKey),
			sendEmoji: (emoji) => sendEmoji(roomKey, emoji),
		}),
		[roomKey],
	);

	return { state, actions };
};
