# 🅿️ Smart Parking System Frontend

## 🌟 Project Overview

Smart Parking System is a modern, user-friendly web application designed to revolutionize parking management. Leveraging cutting-edge web technologies, this application provides an intuitive solution for parking space reservation, visualization, and management.

## 🚀 Key Features

- 🔐 Secure Authentication System
- 📍 Real-time Parking Space Visualization
- 🕒 Reservation Management
- 📱 Responsive and Modern UI
- 🌐 WebSocket Integration for Live Updates

## 🛠 Tech Stack

### Frontend
- **React** (v18.2.0): Core library for building user interfaces
- **React Router** (v6.20.0): Declarative routing
- **Tailwind CSS** (v3.3.5): Utility-first CSS framework
- **Axios**: Promise-based HTTP client
- **Framer Motion**: Animation library
- **React Toastify**: Notification system

### UI Components
- **Headless UI**: Unstyled, accessible UI components
- **Heroicons**: Beautiful hand-crafted SVG icons
- **Material-UI**: Additional React components

### Authentication
- **JWT Decode**: JSON Web Token decoding
- **Secure authentication flow**

## 🏗 Project Structure

```
SmartParkingSystem/
│
├── public/                 # Public assets
├── src/
│   ├── components/         # Reusable React components
│   ├── context/            # React context providers
│   ├── hooks/              # Custom React hooks
│   ├── pages/              # Page components
│   ├── services/           # API and WebSocket services
│   └── styles/             # Global styles
│
├── package.json            # Project dependencies and scripts
└── tailwind.config.js      # Tailwind CSS configuration
```

## 🚦 Getting Started

### Prerequisites
- Node.js (v14+ recommended)
- npm or Yarn

### Installation

1. Clone the repository
```bash
git clone https://github.com/yourusername/SmartParkingSystem.git
cd SmartParkingSystem
```

2. Install dependencies
```bash
npm install
# or
yarn install
```

3. Start the development server
```bash
npm start
# or
yarn start
```

## 📦 Available Scripts

- `npm start`: Runs the app in development mode
- `npm run build`: Builds the app for production
- `npm test`: Launches the test runner
- `npm run eject`: Ejects from Create React App configuration

## 🔍 Browser Support

- Modern browsers (Chrome, Firefox, Safari, Edge)
- Supports browsers with >0.2% market share
- Excludes dead and outdated browsers

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 License

Distributed under the MIT License. See `LICENSE` for more information.

## 📞 Contact

Your Name - [Your Email]

Project Link: [https://github.com/yourusername/SmartParkingSystem](https://github.com/yourusername/SmartParkingSystem)

---

**Built with ❤️ and 🚀 by Your Name** 

## 🗺 Routing Structure

The application uses React Router (v6.20.0) with the following routes:

### Public Routes
- `/login`: Login page
- `/register`: Registration page

### Protected Routes (Inside MainLayout)
- `/`: Home page
- `/reservations`: View and manage parking reservations
- `/profile`: User profile management
- `/parking-lots`: List of available parking lots
- `/parking-lots/:id`: Detailed view of a specific parking lot

### Additional Features
- 🔔 **Toast Notifications**
  - Powered by React Toastify
  - Configurable notifications
  - Position: Top right
  - Auto-close after 5 seconds
  - Draggable and customizable

## 🔒 Authentication Flow

The application implements a comprehensive authentication system with:
- Secure login and registration pages
- JWT-based authentication
- Protected routes
- User profile management

## 🌈 Responsive Design

Built with mobile-first approach using:
- Tailwind CSS for responsive utilities
- Flexbox and Grid layouts
- Adaptive components from Headless UI and Material-UI 