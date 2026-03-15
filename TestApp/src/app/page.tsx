'use client';

import { useState, useEffect } from 'react';
import { createClient, OmniBaseClient } from '@omnibase/omnibase-js';
import { Shield, Key, User as UserIcon, Activity, LogOut, ChevronRight, Zap } from 'lucide-react';

export default function TestPage() {
  const [anonKey, setAnonKey] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [result, setResult] = useState<any>(null);
  const [session, setSession] = useState<any>(null);
  const [client, setClient] = useState<OmniBaseClient | null>(null);

  // Initialize client whenever anonKey changes
  useEffect(() => {
    if (anonKey) {
      const omni = createClient('http://localhost:8000', anonKey);
      setClient(omni);
      
      // Listen for auth changes (exactly like Supabase)
      const { data: { subscription } } = omni.auth.onAuthStateChange((event, session) => {
        console.log('Auth event:', event, session);
        setSession(session);
      });

      return () => subscription.unsubscribe();
    }
  }, [anonKey]);

  const handleSignUp = async () => {
    if (!client) return setResult({ error: 'Please set your Anon Key first' });
    try {
      const { data, error } = await client.auth.signUp({ email, password });
      setResult(error || data);
    } catch (err: any) {
      setResult({ error: err.message });
    }
  };

  const handleSignIn = async () => {
    if (!client) return setResult({ error: 'Please set your Anon Key first' });
    try {
      const { data, error } = await client.auth.signInWithPassword({ email, password });
      setResult(error || data);
    } catch (err: any) {
      setResult({ error: err.message });
    }
  };

  const handleGetUser = async () => {
    if (!client) return;
    try {
      const { data, error } = await client.auth.getUser();
      setResult(error || data);
    } catch (err: any) {
      setResult({ error: err.message });
    }
  };

  const handleSignOut = async () => {
      if(!client) return;
      await client.auth.signOut();
      setResult({ message: 'Signed out successfully' });
  }

  return (
    <div style={{ maxWidth: '1200px', margin: '0 auto', padding: '4rem 2rem', width: '100%' }}>
      <header style={{ marginBottom: '4rem', textAlign: 'center' }}>
        <div style={{ display: 'inline-flex', alignItems: 'center', gap: '0.5rem', background: 'rgba(176, 251, 92, 0.1)', padding: '0.5rem 1rem', borderRadius: '100px', marginBottom: '1.5rem', border: '1px solid rgba(176, 251, 92, 0.2)' }}>
          <Zap size={14} color="var(--brand-primary)" />
          <span style={{ fontSize: '0.875rem', fontWeight: 600, color: 'var(--brand-primary)' }}>OFFICIAL SDK v0.1.0</span>
        </div>
        <h1 className="text-gradient" style={{ fontSize: '3.5rem', fontWeight: 800, marginBottom: '1rem' }}>
          OmniBase TestApp
        </h1>
        <p style={{ color: 'var(--text-secondary)', fontSize: '1.25rem' }}>
          Using <code style={{ color: 'var(--brand-primary)' }}>@omnibase/omnibase-js</code> official library
        </p>
      </header>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '3rem' }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: '2rem' }}>
          <section className="card">
            <h2 style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '1.5rem', fontSize: '1.25rem' }}>
              <Key size={20} color="var(--brand-primary)" />
              SDK Configuration
            </h2>
            <div style={{ marginBottom: '1rem' }}>
              <label style={{ display: 'block', fontSize: '0.875rem', marginBottom: '0.5rem', color: 'var(--text-secondary)' }}>
                Anon Public Key
              </label>
              <input 
                value={anonKey} 
                onChange={(e) => setAnonKey(e.target.value)} 
                placeholder="Paste your OmniBase anon key..."
              />
              <p style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', marginTop: '0.5rem' }}>
                Initializing client with <code style={{ color: '#fff' }}>createClient(URL, KEY)</code>
              </p>
            </div>
          </section>

          <section className="card">
            <h2 style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '1.5rem', fontSize: '1.25rem' }}>
              <Shield size={20} color="var(--brand-primary)" />
              Identity API
            </h2>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              <input 
                value={email} 
                onChange={(e) => setEmail(e.target.value)} 
                type="email" 
                placeholder="Email address"
              />
              <input 
                value={password} 
                onChange={(e) => setPassword(e.target.value)} 
                type="password" 
                placeholder="Password"
              />
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem', marginTop: '0.5rem' }}>
                <button className="btn btn-primary" onClick={handleSignUp}>SignUp User</button>
                <button className="btn btn-secondary" onClick={handleSignIn}>SignIn User</button>
              </div>
            </div>
          </section>

          <section className="card">
            <h2 style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '1.5rem', fontSize: '1.25rem' }}>
              <UserIcon size={20} color="var(--brand-primary)" />
              Authenticated Session
            </h2>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
              <button 
                className="btn btn-secondary" 
                disabled={!session} 
                onClick={handleGetUser}
              >
                Get Account Info
              </button>
              <button className="btn btn-secondary" disabled={!session} onClick={handleSignOut}>
                <LogOut size={16} /> Sign Out
              </button>
            </div>
            {session && (
                <div style={{ marginTop: '1rem', padding: '1rem', background: 'rgba(52, 211, 153, 0.1)', borderRadius: '8px', border: '1px solid rgba(52, 211, 153, 0.2)' }}>
                    <p style={{ fontSize: '0.875rem', color: 'var(--success)', fontWeight: 600 }}>Active Session Found</p>
                    <p style={{ fontSize: '0.75rem', opacity: 0.7 }}>User: {session.user.email}</p>
                </div>
            )}
          </section>
        </div>

        <div style={{ display: 'flex', flexDirection: 'column' }}>
          <section className="card glass" style={{ flex: 1, minHeight: '500px', display: 'flex', flexDirection: 'column', borderColor: result?.error ? 'var(--error)' : 'var(--border-subtle)' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
              <h2 style={{ fontSize: '1.25rem' }}>Response Payload</h2>
              <div style={{ display: 'flex', gap: '0.5rem' }}>
                  <span style={{ padding: '0.25rem 0.5rem', borderRadius: '4px', background: 'rgba(255,255,255,0.05)', fontSize: '0.75rem' }}>JSON</span>
              </div>
            </div>
            <pre style={{ 
              background: 'rgba(0,0,0,0.3)', 
              padding: '1.5rem', 
              borderRadius: '8px', 
              flex: 1, 
              overflow: 'auto',
              fontFamily: 'monospace',
              fontSize: '0.875rem',
              color: result?.error ? '#ff8080' : 'var(--text-primary)'
            }}>
              {result ? JSON.stringify(result, null, 2) : '// Awaiting SDK operations...'}
            </pre>
          </section>
        </div>
      </div>
    </div>
  );
}
