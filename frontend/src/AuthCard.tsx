import React, { useState, useContext } from 'react';
import { Eye, EyeOff, Mail, Lock, ArrowRight, Zap } from 'lucide-react';
import { AuthContext } from "./context/AuthContext";

export default function AuthPages() {
  const { setToken } = useContext(AuthContext);
  const [currentPage, setCurrentPage] = useState('login');
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState<string | null>(null);

  async function handleLogin(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setSuccess(null);
    setLoading(true);
    try {
      const res = await fetch("http://localhost:8080/api/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });
      const data = await res.json();
      if (!res.ok) return setError(data.error || "Login failed");
      setSuccess("Login successful!");
      setToken(data.token);
    } catch {
      setError("Network error. Please try again.");
    } finally {
      setLoading(false);
    }
  }

  async function handleRegister(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setSuccess(null);
    if (password !== confirmPassword) return setError("Passwords do not match");
    if (password.length < 8) return setError("Password must be at least 8 characters");
    setLoading(true);
    try {
      const res = await fetch("http://localhost:8080/api/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });
      const data = await res.json();
      if (!res.ok) return setError(data.error || "Registration failed");
      setSuccess("Account created! Redirecting to login...");
      setTimeout(() => {
        setCurrentPage("login");
        setEmail("");
        setPassword("");
        setConfirmPassword("");
        setSuccess(null);
      }, 2000);
    } catch {
      setError("Network error. Please try again.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="w-screen h-screen min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-950 via-purple-950 to-slate-950 relative overflow-hidden">
      {/* Animated Background Blobs */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div className="absolute top-0 right-0 w-96 h-96 bg-purple-600 rounded-full mix-blend-multiply filter blur-3xl opacity-15 animate-pulse"></div>
        <div className="absolute bottom-0 left-0 w-96 h-96 bg-blue-600 rounded-full mix-blend-multiply filter blur-3xl opacity-15 animate-pulse" style={{animationDelay: '2s'}}></div>
      </div>

      {/* Main Container */}
      <div className="relative z-10 w-full h-full flex items-center justify-center p-4">
        <div className="max-w-5xl w-full h-full md:h-auto">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-0 rounded-3xl overflow-hidden shadow-2xl bg-white/8 backdrop-blur-2xl border border-white/15 h-full md:h-auto">
            
            {/* Left Branding Section */}
            <div className="hidden lg:flex flex-col justify-center p-12 bg-gradient-to-br from-purple-950/60 via-slate-950/80 to-slate-950 text-white border-r border-white/10">
              <div className="space-y-10">
                {/* Logo */}
                <div className="flex items-center gap-3">
                  <div className="w-14 h-14 bg-gradient-to-br from-purple-500 to-blue-600 rounded-xl flex items-center justify-center shadow-lg shadow-purple-500/30 flex-shrink-0">
                    <Zap className="w-7 h-7 text-white" strokeWidth={3} />
                  </div>
                  <div>
                    <h1 className="text-3xl font-bold tracking-tight">KineticOps</h1>
                    <p className="text-xs text-purple-300 font-medium">Automation Platform</p>
                  </div>
                </div>

                {/* Tagline */}
                <div>
                  <h2 className="text-3xl font-bold mb-4 leading-tight">Streamline Your Operations</h2>
                  <p className="text-gray-400 text-sm leading-relaxed">
                    Powerful automation tools designed to <span className="text-purple-300 font-semibold">scale your business</span> effortlessly.
                  </p>
                </div>

                {/* Features */}
                <div className="space-y-5 pt-4">
                  <Feature 
                    label="Real-time Analytics" 
                    text="Track and monitor all operations instantly"
                  />
                  <Feature 
                    label="Secure & Reliable" 
                    text="Enterprise-grade security for your data"
                  />
                  <Feature 
                    label="24/7 Support" 
                    text="Get help whenever you need it"
                  />
                </div>

                {/* Stats */}
                <div className="grid grid-cols-2 gap-6 pt-10 border-t border-white/10">
                  <div>
                    <p className="text-3xl font-bold text-purple-300">10K+</p>
                    <p className="text-xs text-gray-400 mt-2">Active Users</p>
                  </div>
                  <div>
                    <p className="text-3xl font-bold text-blue-300">99.9%</p>
                    <p className="text-xs text-gray-400 mt-2">Uptime</p>
                  </div>
                </div>
              </div>
            </div>

            {/* Right Auth Form Section */}
            <div className="flex flex-col justify-center p-8 lg:p-12 overflow-y-auto md:overflow-y-visible">
              <div className="space-y-6 max-w-md w-full mx-auto">
                
                {/* Tab Switcher */}
                <div className="flex gap-2 bg-white/5 border border-white/10 rounded-xl p-1.5 backdrop-blur-sm">
                  <button
                    onClick={() => { setCurrentPage('login'); setError(null); setSuccess(null); }}
                    className={`flex-1 py-2.5 px-4 rounded-lg font-semibold transition-all duration-300 text-sm ${
                      currentPage === 'login'
                        ? 'bg-gradient-to-r from-purple-600 to-blue-600 text-white shadow-lg shadow-purple-500/30'
                        : 'text-gray-400 hover:text-gray-200'
                    }`}
                  >
                    Login
                  </button>
                  <button
                    onClick={() => { setCurrentPage('signup'); setError(null); setSuccess(null); }}
                    className={`flex-1 py-2.5 px-4 rounded-lg font-semibold transition-all duration-300 text-sm ${
                      currentPage === 'signup'
                        ? 'bg-gradient-to-r from-purple-600 to-blue-600 text-white shadow-lg shadow-purple-500/30'
                        : 'text-gray-400 hover:text-gray-200'
                    }`}
                  >
                    Sign Up
                  </button>
                </div>

                {/* LOGIN FORM */}
                {currentPage === 'login' && (
                  <form onSubmit={handleLogin} className="space-y-4">
                    <div>
                      <h2 className="text-2xl font-bold text-white">Welcome Back</h2>
                      <p className="text-gray-400 text-sm mt-1">Sign in to your account</p>
                    </div>

                    {/* Email Input */}
                    <div className="space-y-2">
                      <label className="block text-xs font-semibold text-gray-300 uppercase tracking-wide">Email</label>
                      <div className="relative group">
                        <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-purple-400 opacity-70 group-focus-within:opacity-100 transition-opacity pointer-events-none" />
                        <input
                          type="email"
                          value={email}
                          onChange={(e) => setEmail(e.target.value)}
                          placeholder="name@company.com"
                          required
                          className="w-full bg-white/5 border border-white/15 text-white placeholder-gray-600 rounded-lg pl-10 pr-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent transition-all hover:bg-white/8"
                        />
                      </div>
                    </div>

                    {/* Password Input */}
                    <div className="space-y-2">
                      <label className="block text-xs font-semibold text-gray-300 uppercase tracking-wide">Password</label>
                      <div className="relative group">
                        <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-purple-400 opacity-70 group-focus-within:opacity-100 transition-opacity pointer-events-none" />
                        <input
                          type={showPassword ? 'text' : 'password'}
                          value={password}
                          onChange={(e) => setPassword(e.target.value)}
                          placeholder="••••••••"
                          required
                          className="w-full bg-white/5 border border-white/15 text-white placeholder-gray-600 rounded-lg pl-10 pr-10 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent transition-all hover:bg-white/8"
                        />
                        <button
                          type="button"
                          tabIndex={-1}
                          onClick={() => setShowPassword(!showPassword)}
                          className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-gray-300 transition-colors p-1"
                        >
                          {showPassword ? <EyeOff size={16} /> : <Eye size={16} />}
                        </button>
                      </div>
                    </div>

                    {/* Remember & Forgot */}
                    <div className="flex items-center justify-between pt-1">
                      <label className="flex items-center gap-2 text-sm text-gray-400 cursor-pointer hover:text-gray-300 transition-colors">
                        <input type="checkbox" className="w-4 h-4 rounded bg-white/10 border border-white/20 accent-purple-500 cursor-pointer" />
                        <span className="text-xs">Remember me</span>
                      </label>
                      <button type="button" className="text-sm text-purple-400 hover:text-purple-300 font-medium transition-colors">
                        Forgot password?
                      </button>
                    </div>

                    {/* Messages */}
                    {error && (
                      <div className="bg-red-500/15 border border-red-500/40 text-red-300 rounded-lg px-4 py-2.5 text-sm">
                        {error}
                      </div>
                    )}
                    {success && (
                      <div className="bg-green-500/15 border border-green-500/40 text-green-300 rounded-lg px-4 py-2.5 text-sm">
                        {success}
                      </div>
                    )}

                    {/* Sign In Button */}
                    <button
                      type="submit"
                      disabled={loading}
                      className="w-full bg-gradient-to-r from-purple-600 to-blue-600 text-white font-semibold py-2.5 rounded-lg hover:shadow-lg hover:shadow-purple-500/40 transition-all duration-300 flex items-center justify-center gap-2 group disabled:opacity-70 disabled:cursor-not-allowed text-sm mt-2"
                    >
                      {loading ? 'Signing in...' : 'Sign In'}
                      {!loading && <ArrowRight size={16} className="group-hover:translate-x-1 transition-transform" />}
                    </button>

                    {/* Sign Up Link */}
                    <p className="text-center text-gray-400 text-xs">
                      Don't have an account?{' '}
                      <button
                        type="button"
                        onClick={() => { setCurrentPage('signup'); setError(null); setSuccess(null); }}
                        className="text-purple-400 hover:text-purple-300 font-semibold transition-colors"
                      >
                        Create one
                      </button>
                    </p>
                  </form>
                )}

                {/* SIGNUP FORM */}
                {currentPage === 'signup' && (
                  <form onSubmit={handleRegister} className="space-y-4">
                    <div>
                      <h2 className="text-2xl font-bold text-white">Create Account</h2>
                      <p className="text-gray-400 text-sm mt-1">Start automating today</p>
                    </div>

                    {/* Email Input */}
                    <div className="space-y-2">
                      <label className="block text-xs font-semibold text-gray-300 uppercase tracking-wide">Email</label>
                      <div className="relative group">
                        <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-purple-400 opacity-70 group-focus-within:opacity-100 transition-opacity pointer-events-none" />
                        <input
                          type="email"
                          value={email}
                          onChange={(e) => setEmail(e.target.value)}
                          placeholder="name@company.com"
                          required
                          className="w-full bg-white/5 border border-white/15 text-white placeholder-gray-600 rounded-lg pl-10 pr-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent transition-all hover:bg-white/8"
                        />
                      </div>
                    </div>

                    {/* Password Input */}
                    <div className="space-y-1.5">
                      <label className="block text-xs font-semibold text-gray-300 uppercase tracking-wide">Password</label>
                      <div className="relative group">
                        <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-purple-400 opacity-70 group-focus-within:opacity-100 transition-opacity pointer-events-none" />
                        <input
                          type={showPassword ? 'text' : 'password'}
                          value={password}
                          onChange={(e) => setPassword(e.target.value)}
                          placeholder="••••••••"
                          required
                          className="w-full bg-white/5 border border-white/15 text-white placeholder-gray-600 rounded-lg pl-10 pr-10 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent transition-all hover:bg-white/8"
                        />
                        <button
                          type="button"
                          tabIndex={-1}
                          onClick={() => setShowPassword(!showPassword)}
                          className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-gray-300 transition-colors"
                        >
                          {showPassword ? <EyeOff size={16} /> : <Eye size={16} />}
                        </button>
                      </div>
                      <p className="text-xs text-gray-400">Minimum 8 characters</p>
                    </div>

                    {/* Confirm Password Input */}
                    <div className="space-y-2">
                      <label className="block text-xs font-semibold text-gray-300 uppercase tracking-wide">Confirm Password</label>
                      <div className="relative group">
                        <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-purple-400 opacity-70 group-focus-within:opacity-100 transition-opacity pointer-events-none" />
                        <input
                          type={showConfirmPassword ? 'text' : 'password'}
                          value={confirmPassword}
                          onChange={(e) => setConfirmPassword(e.target.value)}
                          placeholder="••••••••"
                          required
                          className="w-full bg-white/5 border border-white/15 text-white placeholder-gray-600 rounded-lg pl-10 pr-10 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent transition-all hover:bg-white/8"
                        />
                        <button
                          type="button"
                          tabIndex={-1}
                          onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                          className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-gray-300 transition-colors p-1"
                        >
                          {showConfirmPassword ? <EyeOff size={16} /> : <Eye size={16} />}
                        </button>
                      </div>
                    </div>

                    {/* Terms Checkbox */}
                    <label className="flex items-start gap-2.5 text-xs text-gray-400 cursor-pointer hover:text-gray-300 transition-colors">
                      <input type="checkbox" required className="w-4 h-4 rounded bg-white/10 border border-white/20 accent-purple-500 cursor-pointer mt-0.5 flex-shrink-0" />
                      <span>I agree to the <button type="button" className="text-purple-400 hover:text-purple-300 underline">Terms</button> and <button type="button" className="text-purple-400 hover:text-purple-300 underline">Privacy Policy</button></span>
                    </label>

                    {/* Messages */}
                    {error && (
                      <div className="bg-red-500/15 border border-red-500/40 text-red-300 rounded-lg px-4 py-2.5 text-sm">
                        {error}
                      </div>
                    )}
                    {success && (
                      <div className="bg-green-500/15 border border-green-500/40 text-green-300 rounded-lg px-4 py-2.5 text-sm">
                        {success}
                      </div>
                    )}

                    {/* Sign Up Button */}
                    <button
                      type="submit"
                      disabled={loading}
                      className="w-full bg-gradient-to-r from-purple-600 to-blue-600 text-white font-semibold py-2.5 rounded-lg hover:shadow-lg hover:shadow-purple-500/40 transition-all duration-300 flex items-center justify-center gap-2 group disabled:opacity-70 disabled:cursor-not-allowed text-sm mt-2"
                    >
                      {loading ? 'Creating account...' : 'Create Account'}
                      {!loading && <ArrowRight size={16} className="group-hover:translate-x-1 transition-transform" />}
                    </button>

                    {/* Login Link */}
                    <p className="text-center text-gray-400 text-xs">
                      Already have an account?{' '}
                      <button
                        type="button"
                        onClick={() => { setCurrentPage('login'); setError(null); setSuccess(null); }}
                        className="text-purple-400 hover:text-purple-300 font-semibold transition-colors"
                      >
                        Sign in
                      </button>
                    </p>
                  </form>
                )}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function Feature({ label, text }: { label: string; text: string }) {
  return (
    <div className="flex items-start gap-3">
      <div className="w-8 h-8 bg-gradient-to-br from-purple-500 to-blue-600 rounded-lg flex items-center justify-center flex-shrink-0 shadow-lg shadow-purple-500/30">
        <span className="text-white text-xs font-bold">✓</span>
      </div>
      <div>
        <h3 className="font-semibold text-white text-sm mb-0.5">{label}</h3>
        <p className="text-gray-400 text-xs leading-relaxed">{text}</p>
      </div>
    </div>
  );
}