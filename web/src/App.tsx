import { BrowserRouter, Link, Route, Routes } from "react-router-dom";
import "./App.css";

import { HomePage } from "./pages/HomePage";
import { LoginPage } from "./pages/LoginPage";
import { RegisterPage } from "./pages/RegisterPage";
import { ConfirmationPage } from "./pages/ConfirmationPage";
import { GetUserPage } from "./pages/GetUserPage";
import { PostPage } from "./pages/PostPage";

export const API_URL = "/v1";

function Navigation() {
  return (
    <header className="navbar">
      <div className="navbar-inner">
        <Link className="logo" to="/">
          GopherSocial
        </Link>

        <nav>
          <Link to="/">Home</Link>
          <Link to="/login">Login</Link>
          <Link to="/register">Register</Link>
        </nav>
      </div>
    </header>
  );
}

function App() {
  return (
    <BrowserRouter>
      <Navigation />

      <main className="app-container">
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
          <Route
            path="/activate/:token"
            element={<ConfirmationPage />}
          />
          <Route
            path="/users/:userID"
            element={<GetUserPage />}
          />
          <Route
            path="/posts/:postID"
            element={<PostPage />}
          />
        </Routes>
      </main>
    </BrowserRouter>
  );
}

export default App;