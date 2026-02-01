import { useState, useEffect } from 'react';
import { api } from '../api/client';

function TokenList({ onSelectToken }) {
  const [tokens, setTokens] = useState([]);
  const [loading, setLoading] = useState(false);
  const [syncing, setSyncing] = useState(false);

  const loadTokens = async () => {
    setLoading(true);
    try {
      const data = await api.getAllTokens();
      // Сортируем по market cap (от большего к меньшему)
      const sortedTokens = (data.tokens || []).sort((a, b) => b.market_cap - a.market_cap);
      setTokens(sortedTokens);
    } catch (error) {
      console.error('Failed to load tokens:', error);
    }
    setLoading(false);
  };

  const handleSync = async () => {
    setSyncing(true);
    try {
      await api.syncTokens(10);
      await loadTokens();
    } catch (error) {
      console.error('Sync failed:', error);
    }
    setSyncing(false);
  };

  useEffect(() => {
    loadTokens();
  }, []);

  const formatPrice = (price) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(price);
  };

  const formatMarketCap = (cap) => {
    if (cap >= 1e12) return `$${(cap / 1e12).toFixed(2)}T`;
    if (cap >= 1e9) return `$${(cap / 1e9).toFixed(2)}B`;
    if (cap >= 1e6) return `$${(cap / 1e6).toFixed(2)}M`;
    return formatPrice(cap);
  };

  return (
    <div className="token-list">
      <div className="header">
        <h2>🪙 Crypto Tokens</h2>
        <button onClick={handleSync} disabled={syncing}>
          {syncing ? '⏳ Syncing...' : '🔄 Sync from CoinGecko'}
        </button>
      </div>

      {loading ? (
        <div className="loading">Loading tokens...</div>
      ) : (
        <table>
          <thead>
            <tr>
              <th>Rank</th>
              <th>Symbol</th>
              <th>Name</th>
              <th>Price</th>
              <th>Market Cap</th>
              <th>24h Volume</th>
              <th>Action</th>
            </tr>
          </thead>
          <tbody>
            {tokens.map((token, index) => (
              <tr key={token.id}>
                <td className="rank">#{index + 1}</td>
                <td className="symbol">{token.symbol.toUpperCase()}</td>
                <td>{token.name}</td>
                <td className="price">{formatPrice(token.current_price)}</td>
                <td>{formatMarketCap(token.market_cap)}</td>
                <td>{formatMarketCap(token.volume_24h)}</td>
                <td>
                  <button 
                    className="view-btn"
                    onClick={() => onSelectToken(token.id)}
                  >
                    📊 View
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}

export default TokenList;