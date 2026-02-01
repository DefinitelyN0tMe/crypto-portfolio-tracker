import { useState, useEffect } from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { api } from '../api/client';

function PriceChart({ tokenId }) {
  const [history, setHistory] = useState([]);
  const [tokenInfo, setTokenInfo] = useState(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    loadData();
  }, [tokenId]);

  const loadData = async () => {
    setLoading(true);
    try {
      // Get token info
      const token = await api.getToken(tokenId);
      setTokenInfo(token);

      // Get price history
      const historyData = await api.getPriceHistory(tokenId, 50);
      
      // Format for chart (reverse to show oldest first)
      const chartData = historyData.history
        .reverse()
        .map(point => ({
          time: new Date(point.timestamp).toLocaleTimeString(),
          price: point.price,
        }));

      setHistory(chartData);
    } catch (error) {
      console.error('Failed to load chart data:', error);
    }
    setLoading(false);
  };

  const formatPrice = (price) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 2,
    }).format(price);
  };

  if (loading) return <div className="loading">Loading chart...</div>;
  if (!tokenInfo) return null;

  return (
    <div className="price-chart">
      <div className="chart-header">
        <h2>📈 {tokenInfo.name} ({tokenInfo.symbol.toUpperCase()})</h2>
        <div className="current-price">
          <span className="label">Current Price:</span>
          <span className="value">{formatPrice(tokenInfo.current_price)}</span>
        </div>
      </div>

      {history.length > 0 ? (
        <ResponsiveContainer width="100%" height={400}>
          <LineChart data={history}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis 
              dataKey="time" 
              angle={-45}
              textAnchor="end"
              height={80}
            />
            <YAxis 
              tickFormatter={(value) => `$${value.toLocaleString()}`}
            />
            <Tooltip 
              formatter={(value) => formatPrice(value)}
              labelStyle={{ color: '#000' }}
            />
            <Legend />
            <Line 
              type="monotone" 
              dataKey="price" 
              stroke="#8884d8" 
              strokeWidth={2}
              dot={false}
              name="Price (USD)"
            />
          </LineChart>
        </ResponsiveContainer>
      ) : (
        <div className="no-data">No price history available yet. Worker will collect data soon!</div>
      )}

      <button onClick={loadData} className="refresh-btn">
        🔄 Refresh Data
      </button>
    </div>
  );
}

export default PriceChart;