// For form validation, you can expand this based on your needs

document.getElementById('loginForm').onsubmit = async function(event) {
    event.preventDefault();
    
    const formData = new FormData(this);
    const response = await fetch('/login', {
        method: 'POST',
        body: formData
    });

    if (response.ok) {
        window.location.href = '/theme-dust.html'; // Redirect to home page if login successful
    } else {
        alert('Login failed! Check your credentials.');
    }
};

document.getElementById('registerForm').onsubmit = async function(event) {
    event.preventDefault();
    
    const formData = new FormData(this);
    const response = await fetch('/register', {
        method: 'POST',
        body: formData
    });

    if (response.ok) {
        window.location.href = '/login';  // Redirect to login page if registration successful
    } else {
        alert('Registration failed! Username might be taken.');
    }
};

document.getElementById('registerForm').addEventListener('submit', function(event) {
    event.preventDefault();
    
    // Get the values from the form
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;

    // Check if the user already exists
    if (localStorage.getItem(username)) {
        alert('User already exists!');
    } else {
        // Store the user information in LocalStorage
        localStorage.setItem(username, password);
        alert('Registration successful! You can now log in.');

        // Redirect to login page
        window.location.href = 'login.html';
    }
});

document.getElementById('loginForm').addEventListener('submit', function(event) {
    event.preventDefault();
    
    // Get the username and password from the form
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;

    // Check the stored password in localStorage
    const storedPassword = localStorage.getItem(username);

    if (storedPassword === password) {
        alert('Login successful!');
        // Redirect to home page or dashboard
        window.location.href = 'home.html';
        localStorage.setItem('loggedInUser', username); // Save logged-in user
    } else {
        alert('Incorrect username or password');
    }
});