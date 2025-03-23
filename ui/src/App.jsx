import React from 'react';
import Dashboard from './components/Dashboard';
import { Toaster } from './components/ui/use-toast';

function App() {
  return (
    <>
      <Dashboard />
      <Toaster />
    </>
  );
}

export default App;
