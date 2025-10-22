import { useContext, useEffect, useState } from "react";
import { AuthContext } from "./context/AuthContext";
import { FiLogOut } from "react-icons/fi";
import { Link } from "react-router-dom";

export default function Dashboard() {
  const { token, setToken } = useContext(AuthContext);
  const [profile, setProfile] = useState<{ id: number; email: string } | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    fetch("http://localhost:8080/api/profile", {
      headers: { Authorization: `Bearer ${token}` },
    })
      .then(async (res) => {
        if (!res.ok) {
          setError("Session expired or unauthorized.");
          setProfile(null);
          setLoading(false);
          return;
        }
        const data = await res.json();
        setProfile(data);
        setLoading(false);
      })
      .catch(() => {
        setError("Network error, try again.");
        setProfile(null);
        setLoading(false);
      });
  }, [token]);

  function handleLogout() {
    setToken(null);
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-900 to-blue-900 flex flex-col items-center justify-center px-3">
      <div className="max-w-lg w-full bg-white/10 backdrop-blur-lg shadow-2xl rounded-2xl p-10 flex flex-col items-center border border-white/10">
        <img src="/kineticops-logo.svg" alt="KineticOps Logo" className="w-20 mb-4" />
        <h2 className="text-2xl font-bold text-white mb-4">User Profile</h2>
        {loading && <p className="text-cyan-200 mb-2">Loading profile...</p>}
        {error && <p className="mb-4 text-red-300">{error}</p>}
        {profile && (
          <div className="space-y-2 text-lg text-white text-center mb-8">
            <div>
              <span className="font-semibold text-blue-400">ID:</span> {profile.id}
            </div>
            <div>
              <span className="font-semibold text-blue-400">Email:</span> {profile.email}
            </div>
          </div>
        )}

        <Link to="/workspaces" className="w-full">
  <button className="w-full mb-6 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 px-6 py-3 rounded-lg text-white font-bold text-base shadow-lg transition-all">
    Manage Workspaces
  </button>
</Link>
        <button
          onClick={handleLogout}
          className="flex items-center px-5 py-2 text-sm font-bold rounded bg-gradient-to-r from-blue-500 to-purple-600 text-white shadow hover:from-blue-600 hover:to-purple-700 gap-2">
          <FiLogOut /> Logout
        </button>
      </div>
    </div>
  );
}
