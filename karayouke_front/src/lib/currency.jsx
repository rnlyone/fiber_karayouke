// Currency utility - IDR (Indonesian Rupiah) as primary currency via iPaymu

// Format price in IDR
export const formatIDR = (amount) => {
	try {
		return new Intl.NumberFormat('id-ID', {
			style: 'currency',
			currency: 'IDR',
			minimumFractionDigits: 0,
			maximumFractionDigits: 0,
		}).format(amount);
	} catch {
		return `Rp${Number(amount).toLocaleString('id-ID')}`;
	}
};

// Create a currency context hook (simplified - always IDR)
import { createContext, useContext, useMemo } from 'react';

const CurrencyContext = createContext(null);

export const CurrencyProvider = ({ children }) => {
	const value = useMemo(() => ({
		currency: 'IDR',
		format: (amount) => formatIDR(amount),
		info: { code: 'IDR', symbol: 'Rp', locale: 'id-ID' },
	}), []);

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

export default { formatIDR };
