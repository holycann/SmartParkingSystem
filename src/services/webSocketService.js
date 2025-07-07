// WebSocket connection management
let socket = null;
const handlerMap = new Map();
let reconnectAttempts = 0;
const maxReconnectAttempts = 5;
let messageHandlers = [];
let reconnectTimer = null;

// WebSocket message types
export const WS_MESSAGE_TYPES = {
  NOTIFICATION_UPDATE: 'NOTIFICATION_UPDATE',
  RESERVATION_ADDED: 'RESERVATION_ADD',
  PARKING_UPDATE: 'PARKING_UPDATE',
  GATE_OPEN: 'gate_open',
  GATE_CLOSE: 'gate_close',
};

/**
 * Connect to WebSocket server
 * @param {Function} onMessage - Callback for message events
 * @param {Function} onConnect - Callback when connection is established
 * @param {Function} onDisconnect - Callback when connection is closed
 * @returns {Function} Cleanup function
 */
export const connectWebSocket = (onMessage, onConnect, onDisconnect) => {
  // If a message handler is provided, add it to the handlers array
  if (onMessage && typeof onMessage === 'function') {
    messageHandlers.push(onMessage);
  }

  if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) {
    if (socket.readyState === WebSocket.OPEN) {
      if (onConnect) onConnect();
    } else if (socket.readyState === WebSocket.CONNECTING) {
      const handleOpen = () => {
        if (onConnect) onConnect();
        socket.removeEventListener('open', handleOpen);
      };
      socket.addEventListener('open', handleOpen);
    }

    return () => {
      messageHandlers = messageHandlers.filter(handler => handler !== onMessage);
    };
  }

  //  Get the token for authentication
  const token = localStorage.getItem('token');
  if (!token) {
    console.error('No authentication token found for WebSocket connection');
    if (onDisconnect) onDisconnect();
    return () => { };
  }

  // Determine WebSocket URL
  const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const wsBaseUrl = process.env.NEXT_PUBLIC_WS_BASE_URL || `${wsProtocol}//${window.location.hostname}:8080`;
  const wsUrl = `${wsBaseUrl}/ws?token=${token}`;

  try {
    // Create WebSocket connection
    socket = new WebSocket(wsUrl);

    // Connection opened
    socket.addEventListener('open', () => {
      console.log('WebSocket connection established');
      reconnectAttempts = 0;
      if (onConnect) onConnect();
    });

    // Listen for messages
    socket.addEventListener('message', (event) => {
      // Call all registered message handlers
      let parsedData;
      try {
        parsedData = JSON.parse(event.data);
      } catch (err) {
        console.error('Failed to parse WebSocket message:', err);
        return;
      }

      console.log(parsedData)

      const parsedType = parsedData.type || '';

      // ===== Global Handler =====
      if (handlerMap.has('')) {
        for (const handler of handlerMap.get('')) {
          try {
            handler(parsedData);
          } catch (err) {
            console.error('Error in global WS handler:', err);
          }
        }
      }

      // ===== Prefix Handlers =====
      const matchedPrefixes = []; // <== Track prefix yang match

      for (const [prefix, handlers] of handlerMap.entries()) {
        if (prefix !== '' && parsedType.startsWith(prefix)) {
          matchedPrefixes.push(prefix); // Catat prefix yang match
          for (const handler of handlers) {
            try {
              handler(parsedData);
            } catch (err) {
              console.error(`Error in handler for prefix "${prefix}":`, err);
            }
          }
        }
      }

      // ===== Check Prefix Match =====
      if (matchedPrefixes.length === 0) {
        const availablePrefixes = Array.from(handlerMap.keys()).filter(p => p !== '');
        console.warn(`No prefix handler matched for type "${parsedType}". Available prefixes:`, availablePrefixes);
      }

      // ===== General Handlers =====
      messageHandlers.forEach(handler => {
        try {
          handler(parsedData);
        } catch (err) {
          console.error('Error in general WS message handler:', err);
        }
      });
    });

    // Connection closed
    socket.addEventListener('close', (event) => {
      console.log(`WebSocket connection closed. Code: ${event.code}, Reason: ${event.reason}`);
      socket = null;

      // Call the disconnect callback
      if (onDisconnect) onDisconnect();

      // Attempt to reconnect if not a normal closure
      if (event.code !== 1000 && event.code !== 1001) {
        handleReconnect();
      }
    });

    // Connection error
    socket.addEventListener('error', (error) => {
      console.error('WebSocket error:', error);
    });

    // Return cleanup function
    return () => {
      // Remove the specific message handler
      messageHandlers = messageHandlers.filter(handler => handler !== onMessage);
    };
  } catch (error) {
    console.error('Error creating WebSocket connection:', error);
    if (onDisconnect) onDisconnect();
    return () => { };
  }
};

/**
 * Handle WebSocket reconnection
 */
const handleReconnect = () => {
  // Clear any existing reconnect timer
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
  }

  // Check if we've exceeded the maximum reconnect attempts
  if (reconnectAttempts >= maxReconnectAttempts) {
    console.error(`Maximum reconnect attempts (${maxReconnectAttempts}) reached. Giving up.`);
    return;
  }

  // Exponential backoff for reconnect
  const delay = Math.min(1000 * Math.pow(2, reconnectAttempts), 30000);
  console.log(`Attempting to reconnect in ${delay}ms (attempt ${reconnectAttempts + 1}/${maxReconnectAttempts})`);

  reconnectTimer = setTimeout(() => {
    reconnectAttempts++;

    // Attempt to reconnect with the same handlers
    connectWebSocket(
      null, // Don't add duplicate message handlers
      () => console.log('WebSocket reconnected successfully'),
      () => console.log('WebSocket reconnection failed')
    );
  }, delay);
};

/**
 * Disconnect from WebSocket server
 */
export const disconnectWebSocket = () => {
  if (socket) {
    socket.close(1000, 'User initiated disconnect');
    socket = null;
  }

  // Clear any reconnect timer
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }

  // Clear all handlers
  messageHandlers = [];
};

/**
 * Send a message through the WebSocket connection
 * @param {string} message - Message to send
 * @returns {boolean} Success status
 */
export const sendWebSocketMessage = (message) => {
  if (!socket || socket.readyState !== WebSocket.OPEN) {
    console.error('Cannot send message: WebSocket is not connected');
    return false;
  }

  try {
    socket.send(message);
    return true;
  } catch (error) {
    console.error('Error sending WebSocket message:', error);
    return false;
  }
};

export const registerHandlerByPrefix = (prefix, handler) => {
  if (!handlerMap.has(prefix)) {
    handlerMap.set(prefix, []);
  }
  handlerMap.get(prefix).push(handler);

  // Return unregister function
  return () => {
    handlerMap.set(prefix, handlerMap.get(prefix).filter(h => h !== handler));
  };
};

// Export a WebSocket status check function
export const isWebSocketConnected = () => {
  return socket && socket.readyState === WebSocket.OPEN;
};