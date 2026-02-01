import axios from 'axios';

const API_BASE_URL = 'http://localhost:8080/api/v1';

export const api = {
  // Get all tokens via search
  searchTokens: async (query = 'bitcoin') => {
    const response = await axios.get(`${API_BASE_URL}/search?q=${query}`);
    return response.data;
  },

  // Get token by ID
  getToken: async (id) => {
    const response = await axios.get(`${API_BASE_URL}/tokens/${id}`);
    return response.data;
  },

  // Sync tokens from CoinGecko
  syncTokens: async (limit = 10) => {
    const response = await axios.post(`${API_BASE_URL}/sync?limit=${limit}`);
    return response.data;
  },

  // Get price history
  getPriceHistory: async (id, limit = 50) => {
    const response = await axios.get(`${API_BASE_URL}/history/${id}?limit=${limit}`);
    return response.data;
  },

  // Get analytics
  getAnalytics: async () => {
    const response = await axios.get(`${API_BASE_URL}/analytics`);
    return response.data;
  },

  // Get all tokens
  getAllTokens: async () => {
    const response = await axios.get(`${API_BASE_URL}/tokens`);
    return response.data;
  }
};