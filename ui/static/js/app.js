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

// Screen Wake Lock API Helpers
let globalWakeLock = null;

async function requestScreenWakeLock() {
  if ('wakeLock' in navigator) {
    try {
      globalWakeLock = await navigator.wakeLock.request('screen');
      console.log('⚡ Screen Wake Lock acquired and active!');
      globalWakeLock.addEventListener('release', () => {
        console.log('⚡ Screen Wake Lock was released');
      });
    } catch (err) {
      console.warn('⚠️ Screen Wake Lock request failed:', err);
    }
  } else {
    console.warn('⚠️ Screen Wake Lock API is not supported in this browser.');
  }
}

function releaseScreenWakeLock() {
  if (globalWakeLock) {
    globalWakeLock.release()
      .then(() => {
        globalWakeLock = null;
        console.log('⚡ Screen Wake Lock released successfully');
      })
      .catch(err => {
        console.error('❌ Failed to release Screen Wake Lock:', err);
      });
  }
}

// Automatically re-acquire Wake Lock when tab returns to focus
document.addEventListener('visibilitychange', async () => {
  if (globalWakeLock !== null && document.visibilityState === 'visible') {
    await requestScreenWakeLock();
  }
});

// Silent Base64 Audio Keep-Alive Loop (tricks OS/browser sandbox to keep JS event loop warm)
let globalSilentAudio = null;

function startBackgroundKeepAlive() {
  if (!globalSilentAudio) {
    // Minimal silent WAV audio file loop
    globalSilentAudio = new Audio('data:audio/wav;base64,UklGRigAAABXQVZFZm10IBIAAAABAAEARKwAAIhYAQACABAAAABkYXRhAgAAAAEA');
    globalSilentAudio.loop = true;
  }
  
  globalSilentAudio.play()
    .then(() => {
      console.log('⚡ Background audio keep-alive active! Tab suspension prevented.');
    })
    .catch(err => {
      console.warn('⚠️ Audio keep-alive loop waiting for user interaction to play:', err);
    });
}

function stopBackgroundKeepAlive() {
  if (globalSilentAudio) {
    globalSilentAudio.pause();
    console.log('⚡ Background audio keep-alive deactivated.');
  }
}
