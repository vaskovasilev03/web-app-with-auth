document.addEventListener("DOMContentLoaded", () => {
    if (document.getElementById('register-form')) {
        loadCaptcha();
    }
    fetch("/api/session", { headers: { "Accept": "application/json" } })
        .then(response => response.json())
        .then(data => {
            if (data.authenticated) {
                const welcomeMsg = document.getElementById("welcome-msg");
                welcomeMsg.innerText = `Welcome, ${data.firstName} ${data.lastName}!`;
                
                document.getElementById("auth-links").style.display = "none";
                document.getElementById("user-actions").style.display = "block";
            }
        })
        .catch(err => console.error("Session check failed:", err));
});

function logout() {
    fetch("/logout", { method: "POST" }).then(() => {
        window.location.reload();
    });
}

const loginForm = document.getElementById('login-form');
if (loginForm) {
    loginForm.addEventListener('submit', async (e) => {
        e.preventDefault();

        const email = document.getElementById('email').value;
        const password = document.getElementById('password').value;
        const errorMsg = document.getElementById('error-message');

        try {
            const response = await fetch('/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ email, password })
            });

            if (response.ok) {
                window.location.href = '/';
            } else {
                const data = await response.text();
                errorMsg.innerText = data || "Invalid credentials";
                errorMsg.style.display = 'block';
            }
        } catch (err) {
            errorMsg.innerText = "Connection failed. Is the server running?";
            errorMsg.style.display = 'block';
        }
    });
}

async function loadCaptcha() {
    try {
        const response = await fetch('/captcha');
        const data = await response.json();
        
        const equation = `${data.num1} ${data.operator} ${data.num2}`;
        
        document.getElementById('captcha-equation').innerText = equation;
        document.getElementById('captcha-id').value = data.captcha_id;
    } catch (err) {
        console.error("Failed to load captcha", err);
    }
}

const registerForm = document.getElementById('register-form');
if (registerForm) {
    registerForm.addEventListener('submit', async (e) => {
        e.preventDefault();

        const payload = {
            first_name: document.getElementById('firstName').value,
            last_name: document.getElementById('lastName').value,
            email: document.getElementById('email').value,
            password: document.getElementById('password').value,
            captcha_id: document.getElementById('captcha-id').value,
            captcha_answer: document.getElementById('captcha-answer').value
        };

        const errorMsg = document.getElementById('error-message');

        try {
            const response = await fetch('/register', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            });

            if (response.ok) {
                window.location.href = '/';
            } else {
                const text = await response.text();
                errorMsg.innerText = text;
                errorMsg.style.display = 'block';
                loadCaptcha();
            }
        } catch (err) {
            errorMsg.innerText = "Connection error";
            errorMsg.style.display = 'block';
        }
    });
}