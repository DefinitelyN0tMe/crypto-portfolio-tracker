import { useState, useEffect } from 'react';
import { api } from '../api/client';

function Analytics() {
  const [analytics, setAnalytics] = useState(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    loadAnalytics();
  }, []);

  const loadAnalytics = async () => {
    setLoading(true);
    try {
      const data = await api.getAnalytics();
      setAnalytics(data.analytics);
    } catch (error) {
      console.error('Failed to load analytics:', error);
    }
    setLoading(false);
  };

  const formatLargeNumber = (num) => {
    if (num >= 1e12) return `$${(num / 1e12).toFixed(2)}T`;
    if (num >= 1e9) return `$${(num / 1e9).toFixed(2)}B`;
    if (num >= 1e6) return `$${(num / 1e6).toFixed(2)}M`;
    return `$${num.toFixed(2)}`;
  };

  if (loading) return <div className="loading">Loading analytics...</div>;
  if (!analytics) return null;

  return (
    <div className="analytics">
      <h2>📊 Market Analytics</h2>
      
      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-label">💰 Total Market Cap</div>
          <div className="stat-value">
            {formatLargeNumber(analytics.total_market_cap?.value || 0)}
          </div>
        </div>

        <div className="stat-card">
          <div className="stat-label">📈 Average Token Price</div>
          <div className="stat-value">
            ${(analytics.avg_price?.value || 0).toFixed(2)}
          </div>
        </div>

        <div className="stat-card">
          <div className="stat-label">🪙 Top Tokens</div>
          <div className="stat-value">
            {analytics.top_tokens?.buckets?.length || 0}
          </div>
        </div>
      </div>

      <div className="top-tokens">
        <h3>🏆 Top 10 Tokens by Market Cap</h3>
        <div className="tokens-grid">
          {analytics.top_tokens?.buckets?.map((token, index) => (
            <div key={token.key} className="token-card">
              <div className="rank">#{index + 1}</div>
              <div className="token-symbol">{token.key.toUpperCase()}</div>
              <div className="token-cap">
                {formatLargeNumber(token.by_market_cap.value)}
              </div>
            </div>
          ))}
        </div>
      </div>

      <button onClick={loadAnalytics} className="refresh-btn">
        🔄 Refresh Analytics
      </button>
    </div>
  );
}

export default Analytics;