import { useState, useEffect, createContext, useContext } from 'react';

const AUTH_TOKEN_KEY = 'karayouke:auth:token';
const AUTH_USER_KEY = 'karayouke:auth:user';

const API_BASE = (() => {
	const raw = import.meta.env.VITE_WS_HOST?.trim();
	if (raw && raw.length > 0) {
		return raw.replace(/\/$/, '');
	}
	if (typeof window !== 'undefined' && window.location) {
		return window.location.origin.replace(/\/$/, '');
	}
	return '';
})();

const AuthContext = createContext(null);

export const useAuth = () => {
	const context = useContext(AuthContext);
	if (!context) {
		throw new Error('useAuth must be used within an AuthProvider');
	}
	return context;
};

const getStoredAuth = () => {
	const token = localStorage.getItem(AUTH_TOKEN_KEY);
	const userStr = localStorage.getItem(AUTH_USER_KEY);
	if (token && userStr) {
		try {
			return { token, user: JSON.parse(userStr) };
		} catch {
			return null;
		}
	}
	return null;
};

const storeAuth = (token, user) => {
	localStorage.setItem(AUTH_TOKEN_KEY, token);
	localStorage.setItem(AUTH_USER_KEY, JSON.stringify(user));
};

const clearAuth = () => {
	localStorage.removeItem(AUTH_TOKEN_KEY);
	localStorage.removeItem(AUTH_USER_KEY);
};

export const AuthProvider = ({ children }) => {
	// Use lazy initialization to avoid setState in useEffect
	const [user, setUser] = useState(() => {
		const stored = getStoredAuth();
		return stored?.user || null;
	});
	const [token, setToken] = useState(() => {
		const stored = getStoredAuth();
		return stored?.token || null;
	});
	const [isLoading, setIsLoading] = useState(true);

	const verifyTokenAsync = async (authToken) => {
		try {
			const response = await fetch(`${API_BASE}/api/auth/me`, {
				headers: { Authorization: `Bearer ${authToken}` },
			});
			if (response.ok) {
				const userData = await response.json();
				setUser(userData);
				storeAuth(authToken, userData);
				return true;
			}
			return false;
		} catch {
			return false;
		}
	};

	useEffect(() => {
		const stored = getStoredAuth();
		if (stored) {
			// Verify token is still valid
			verifyTokenAsync(stored.token).then((valid) => {
				if (!valid) {
					clearAuth();
					setToken(null);
					setUser(null);
				}
				setIsLoading(false);
			});
		} else {
			setIsLoading(false);
		}
	}, []);

	const login = async (email, password) => {
		const response = await fetch(`${API_BASE}/api/auth/login`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ email, password }),
		});

		if (!response.ok) {
			const error = await response.json();
			throw new Error(error.error || 'Login failed');
		}

		const data = await response.json();
		setToken(data.token);
		setUser(data.user);
		storeAuth(data.token, data.user);
		return data.user;
	};

	const register = async (name, username, email, password) => {
		const response = await fetch(`${API_BASE}/api/auth/register`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ name, username, email, password }),
		});

		if (!response.ok) {
			const error = await response.json();
			throw new Error(error.error || 'Registration failed');
		}

		const data = await response.json();
		setToken(data.token);
		setUser(data.user);
		storeAuth(data.token, data.user);
		return data.user;
	};

	const logout = async () => {
		if (token) {
			try {
				await fetch(`${API_BASE}/api/auth/logout`, {
					method: 'POST',
					headers: { Authorization: `Bearer ${token}` },
				});
			} catch {
				// Ignore logout errors
			}
		}
		clearAuth();
		setToken(null);
		setUser(null);
	};

	const value = {
		user,
		token,
		isLoading,
		isAuthenticated: !!user,
		login,
		register,
		logout,
	};

	return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export const getAuthToken = () => localStorage.getItem(AUTH_TOKEN_KEY);

export const fetchWithAuth = async (url, options = {}) => {
	const token = getAuthToken();
	const headers = {
		...options.headers,
	};
	if (token) {
		headers.Authorization = `Bearer ${token}`;
	}
	return fetch(url, { ...options, headers });
};
