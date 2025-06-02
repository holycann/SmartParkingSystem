// src/contexts/WebSocketContext.js
import React, { createContext, useContext, useEffect, useState } from 'react';
import { connectWebSocket } from '../services/webSocketService';
import { useAuth } from './AuthContext';

const WebSocketContext = createContext();

export function WebSocketProvider({ children }) {
    const [wsConnected, setWsConnected] = useState(false);
    const [reconnectAttempt, setReconnectAttempt] = useState(0);
    const { isAuthenticated } = useAuth(); // Import your auth context or method

    // Connect and reconnect logic
    useEffect(() => {
        let cleanup = () => { };

        if (isAuthenticated) {
            console.log('Setting up WebSocket connection');

            const handleConnect = () => {
                setWsConnected(true);
                setReconnectAttempt(0);
                console.log('WebSocket connected');
            };

            const handleDisconnect = () => {
                setWsConnected(false);
                console.log('WebSocket disconnected');

                // Try to reconnect after a delay
                const nextAttempt = reconnectAttempt + 1;
                setReconnectAttempt(nextAttempt);

                // Exponential backoff with max of 30 seconds
                const delay = Math.min(Math.pow(2, nextAttempt) * 1000, 30000);

                setTimeout(() => {
                    if (isAuthenticated) {
                        console.log(`Attempting to reconnect WebSocket (attempt ${nextAttempt})...`);
                        connectWebSocket(null, handleConnect, handleDisconnect);
                    }
                }, delay);
            };

            cleanup = connectWebSocket(null, handleConnect, handleDisconnect);
        }

        return () => {
            console.log('Cleaning up WebSocket connection');
            cleanup();
        };
    }, [isAuthenticated, reconnectAttempt]);

    const contextValue = {
        wsConnected,
        // No need to expose connectWebSocket since it's managed internally
    };

    return (
        <WebSocketContext.Provider value={contextValue}>
            {children}
        </WebSocketContext.Provider>
    );
}

export function useWebSocket() {
    const context = useContext(WebSocketContext);
    if (context === undefined) {
        throw new Error('useWebSocket must be used within a WebSocketProvider');
    }
    return context;
}