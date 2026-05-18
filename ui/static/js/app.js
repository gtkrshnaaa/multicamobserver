// Register Service Worker for PWA
if ('serviceWorker' in navigator) {
  window.addEventListener('load', () => {
    navigator.serviceWorker.register('/static/js/sw.js')
      .then((reg) => {
        console.log('⚡ MulticamObserver Service Worker registered successfully:', reg.scope);
      })
      .catch((err) => {
        console.error('❌ Service Worker registration failed:', err);
      });
  });
}

// PWA Installation Trigger Logic
window.addEventListener('beforeinstallprompt', (e) => {
  // Prevent Chrome 67 and earlier from automatically showing the prompt
  e.preventDefault();
  // Stash the event so it can be triggered later.
  window.deferredPrompt = e;
  
  // Show the install buttons
  const installButtons = document.querySelectorAll('.pwa-install-btn');
  installButtons.forEach(btn => {
    btn.style.display = 'inline-flex';
  });
});

window.addEventListener('appinstalled', (evt) => {
  console.log('⚡ MulticamObserver PWA has been installed successfully!');
  // Hide all install buttons
  const installButtons = document.querySelectorAll('.pwa-install-btn');
  installButtons.forEach(btn => {
    btn.style.display = 'none';
  });
  window.deferredPrompt = null;
});

// Expose prompt trigger function to elements
function triggerPWAInstallation() {
  const promptEvent = window.deferredPrompt;
  if (!promptEvent) return;
  
  // Show the install prompt
  promptEvent.prompt();
  
  // Wait for the user to respond to the prompt
  promptEvent.userChoice.then((choiceResult) => {
    if (choiceResult.outcome === 'accepted') {
      console.log('User accepted the PWA install prompt');
    } else {
      console.log('User dismissed the PWA install prompt');
    }
    window.deferredPrompt = null;
    
    // Hide all install buttons
    const installButtons = document.querySelectorAll('.pwa-install-btn');
    installButtons.forEach(btn => {
      btn.style.display = 'none';
    });
  });
}
