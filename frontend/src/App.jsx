import { useState } from 'react';
import TokenList from './components/TokenList';
import PriceChart from './components/PriceChart';
import Analytics from './components/Analytics';
import './App.css';

function App() {
  const [selectedToken, setSelectedToken] = useState('bitcoin');
  const [activeTab, setActiveTab] = useState('tokens');

  return (
    <div className="app">
      <header className="app-header">
        <h1>🚀 Crypto Portfolio Tracker</h1>
        <p>Real-time cryptocurrency tracking powered by Go, ScyllaDB & ElasticSearch</p>
      </header>

      <nav className="nav-tabs">
        <button 
          className={activeTab === 'tokens' ? 'active' : ''}
          onClick={() => setActiveTab('tokens')}
        >
          🪙 Tokens
        </button>
        <button 
          className={activeTab === 'chart' ? 'active' : ''}
          onClick={() => setActiveTab('chart')}
        >
          📈 Price Chart
        </button>
        <button 
          className={activeTab === 'analytics' ? 'active' : ''}
          onClick={() => setActiveTab('analytics')}
        >
          📊 Analytics
        </button>
      </nav>

      <main className="app-content">
        <div className="content-wrapper">
          {activeTab === 'tokens' && (
            <TokenList onSelectToken={(id) => {
              setSelectedToken(id);
              setActiveTab('chart');
            }} />
          )}

          {activeTab === 'chart' && (
            <PriceChart tokenId={selectedToken} />
          )}

          {activeTab === 'analytics' && (
            <Analytics />
          )}
        </div>
      </main>

      <footer className="app-footer">
        <p>Built with ❤️ using Go, ScyllaDB, ElasticSearch, React & Docker</p>
      </footer>
    </div>
  );
}

export default App;