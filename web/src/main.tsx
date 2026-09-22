import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'
import { createBrowserRouter, RouterProvider } from 'react-router-dom'
import { ConfirmationPage } from './ConfirmationPage.tsx'
import { CreatePostForm } from './CreatePostForm.tsx'
import { GetUserPage } from './GetUserPage.tsx'

const router = createBrowserRouter([
  {
    path: "/",
    element: <App />
  },
  {
    path: "/confirm/:token",
    element: <ConfirmationPage />
  },
  {
    path: "/test",
    element: <CreatePostForm />
  },
  {
    path: "/users/:userID",
    element: <GetUserPage />
  }
])

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <RouterProvider router={router} />
  </StrictMode>,
)
