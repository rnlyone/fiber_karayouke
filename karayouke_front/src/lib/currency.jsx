// Currency utility for locale-based pricing

const EXCHANGE_RATES = {
	USD: 1,
	IDR: 15800, // 1 USD = 15,800 IDR (approximate)
	JPY: 149, // 1 USD = 149 JPY (approximate)
};

const CURRENCY_SYMBOLS = {
	USD: '$',
	IDR: 'Rp',
	JPY: '¥',
};

const CURRENCY_LOCALES = {
	USD: 'en-US',
	IDR: 'id-ID',
	JPY: 'ja-JP',
};

// Detect user's currency based on timezone/locale
export const detectUserCurrency = () => {
	try {
		const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone || '';
		const locale = navigator.language || navigator.userLanguage || '';

		// Check for Japan
		if (
			timezone.includes('Tokyo') ||
			timezone.includes('Japan') ||
			locale.startsWith('ja')
		) {
			return 'JPY';
		}

		// Check for Indonesia
		if (
			timezone.includes('Jakarta') ||
			timezone.includes('Makassar') ||
			timezone.includes('Jayapura') ||
			locale.startsWith('id')
		) {
			return 'IDR';
		}

		// Default to USD for all other locations
		return 'USD';
	} catch {
		return 'USD';
	}
};

// Convert price from USD (base) to target currency
export const convertPrice = (priceInUSD, targetCurrency = 'USD') => {
	const rate = EXCHANGE_RATES[targetCurrency] || 1;
	return Math.round(priceInUSD * rate);
};

// Format price with currency symbol
export const formatPrice = (amount, currency = 'USD') => {
	const locale = CURRENCY_LOCALES[currency] || 'en-US';
	
	try {
		return new Intl.NumberFormat(locale, {
			style: 'currency',
			currency: currency,
			minimumFractionDigits: currency === 'JPY' ? 0 : (currency === 'IDR' ? 0 : 2),
			maximumFractionDigits: currency === 'JPY' ? 0 : (currency === 'IDR' ? 0 : 2),
		}).format(amount);
	} catch {
		const symbol = CURRENCY_SYMBOLS[currency] || '$';
		return `${symbol}${amount.toLocaleString()}`;
	}
};

// Get currency info
export const getCurrencyInfo = (currency = 'USD') => ({
	code: currency,
	symbol: CURRENCY_SYMBOLS[currency] || '$',
	rate: EXCHANGE_RATES[currency] || 1,
	locale: CURRENCY_LOCALES[currency] || 'en-US',
});

// Create a currency context hook
import { createContext, useContext, useState, useMemo } from 'react';

const CurrencyContext = createContext(null);

export const CurrencyProvider = ({ children }) => {
	// Detect currency once during initialization (no useEffect needed)
	const [currency, setCurrency] = useState(() => detectUserCurrency());

	const value = useMemo(() => ({
		currency,
		setCurrency,
		convert: (priceInUSD) => convertPrice(priceInUSD, currency),
		format: (amount) => formatPrice(amount, currency),
		formatFromUSD: (priceInUSD) => formatPrice(convertPrice(priceInUSD, currency), currency),
		info: getCurrencyInfo(currency),
	}), [currency]);

	return (
		<CurrencyContext.Provider value={value}>
			{children}
		</CurrencyContext.Provider>
	);
};

export const useCurrency = () => {
	const context = useContext(CurrencyContext);
	if (!context) {
		throw new Error('useCurrency must be used within a CurrencyProvider');
	}
	return context;
};

export default {
	detectUserCurrency,
	convertPrice,
	formatPrice,
	getCurrencyInfo,
	EXCHANGE_RATES,
	CURRENCY_SYMBOLS,
};
