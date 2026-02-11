import { Navigate, Route, Routes } from 'react-router-dom';
import { AuthProvider } from './lib/auth.jsx';
import { CurrencyProvider } from './lib/currency.jsx';
import './App.css';
import Dashboard from './pages/Dashboard.jsx';
import RoomMaster from './pages/RoomMaster.jsx';
import RoomController from './pages/RoomController.jsx';
import GuestWelcome from './pages/GuestWelcome.jsx';
import RoomPlayer from './pages/RoomPlayer.jsx';
import Login from './pages/Login.jsx';
import Register from './pages/Register.jsx';
import JoinRoom from './pages/JoinRoom.jsx';
import Packages from './pages/Packages.jsx';
import Checkout from './pages/Checkout.jsx';
import PaymentStatus from './pages/PaymentStatus.jsx';
import PaymentHistory from './pages/PaymentHistory.jsx';
import TVPage from './pages/TVPage.jsx';
import TVConnect from './pages/TVConnect.jsx';
import FAQ from './pages/FAQ.jsx';
import RefundPolicy from './pages/RefundPolicy.jsx';
import TermsAndConditions from './pages/TermsAndConditions.jsx';
import Credits from './pages/Credits.jsx';
import {
	AdminDashboard,
	AdminSettings,
	AdminSubscriptionPlans,
	AdminPackages,
	AdminUsers,
	AdminTransactions,
	AdminRooms,
} from './pages/admin/index.js';

function App() {
	return (
		<AuthProvider>
			<CurrencyProvider>
				<Routes>
					<Route path="/" element={<Dashboard />} />
					<Route path="/join" element={<JoinRoom />} />
					<Route path="/tv" element={<TVPage />} />
					<Route path="/tv/connect/:token" element={<TVConnect />} />
					<Route path="/rooms/:roomKey" element={<RoomMaster />} />
					<Route path="/rooms/:roomKey/controller" element={<RoomController />} />
					<Route path="/rooms/:roomKey/guest" element={<GuestWelcome />} />
					<Route path="/rooms/:roomKey/guest/controller" element={<RoomController />} />
					<Route path="/rooms/:roomKey/player" element={<RoomPlayer />} />
					<Route path="/login" element={<Login />} />
					<Route path="/register" element={<Register />} />
					{/* Payment routes */}
					<Route path="/packages" element={<Packages />} />
					<Route path="/checkout/:packageId" element={<Checkout />} />
					<Route path="/payment/status/:transactionId" element={<PaymentStatus />} />
					<Route path="/payment/history" element={<PaymentHistory />} />
					{/* Admin routes */}
					<Route path="/admin" element={<AdminDashboard />} />
					<Route path="/admin/settings" element={<AdminSettings />} />
					<Route path="/admin/subscription-plans" element={<AdminSubscriptionPlans />} />
					<Route path="/admin/packages" element={<AdminPackages />} />
					<Route path="/admin/users" element={<AdminUsers />} />
					<Route path="/admin/transactions" element={<AdminTransactions />} />
					<Route path="/admin/rooms" element={<AdminRooms />} />
					{/* Legal & info pages */}
					<Route path="/faq" element={<FAQ />} />
					<Route path="/refund-policy" element={<RefundPolicy />} />
					<Route path="/terms" element={<TermsAndConditions />} />
					<Route path="/credits" element={<Credits />} />
					<Route path="*" element={<Navigate to="/" replace />} />
				</Routes>
			</CurrencyProvider>
		</AuthProvider>
	);
}

export default App;
