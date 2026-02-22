document.addEventListener("DOMContentLoaded", () => {
    if (document.getElementById('register-form')) {
        loadCaptcha();
    }
    if (window.location.pathname === '/profile') {
        loadProfileData();
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

function loadProfileData() {
    fetch("/api/session")
        .then(res => res.json())
        .then(data => {
            if (data.authenticated) {
                document.getElementById('display-firstname').innerText = data.firstName;
                document.getElementById('display-lastname').innerText = data.lastName;
                document.getElementById('edit-firstname').value = data.firstName;
                document.getElementById('edit-lastname').value = data.lastName;
            } else {
                window.location.href = '/login';
            }
        });
}

function toggleNameEdit() {
    const view = document.getElementById('profile-view');
    const form = document.getElementById('profile-edit-form');
    if (view.style.display === 'none') {
        view.style.display = 'block';
        form.style.display = 'none';
    } else {
        view.style.display = 'none';
        form.style.display = 'block';
    }
}

const profileEditForm = document.getElementById('profile-edit-form');
if (profileEditForm) {
    profileEditForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const firstName = document.getElementById('edit-firstname').value;
        const lastName = document.getElementById('edit-lastname').value;
        const msg = document.getElementById('profile-message');

        try {
            const res = await fetch('/profile/updateName', {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ first_name: firstName, last_name: lastName })
            });
            if (res.ok) {
                msg.innerText = "Name updated successfully!";
                msg.style.color = "green";
                msg.style.display = "block";
                loadProfileData();
                toggleNameEdit();
            } else {
                msg.innerText = await res.text();
                msg.style.color = "red";
                msg.style.display = "block";
            }
        } catch (err) {
            console.error(err);
        }
    });
}

function togglePasswordEdit() {
    const btn = document.getElementById('change-password-btn');
    const form = document.getElementById('password-edit-form');
    if (form.style.display === 'none') {
        form.style.display = 'block';
        btn.style.display = 'none';
    } else {
        form.style.display = 'none';
        btn.style.display = 'block';
        document.getElementById('password-edit-form').reset();
    }
}

const passwordEditForm = document.getElementById('password-edit-form');
if (passwordEditForm) {
    passwordEditForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const currentPassword = document.getElementById('current-password').value;
        const newPassword = document.getElementById('new-password').value;
        const repeatPassword = document.getElementById('repeat-password').value;
        const msg = document.getElementById('profile-message');

        if (newPassword !== repeatPassword) {
            msg.innerText = "New passwords do not match";
            msg.style.color = "red";
            msg.style.display = "block";
            return;
        }

        try {
            const res = await fetch('/profile/updatePassword', {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ current_password: currentPassword, new_password: newPassword })
            });
            if (res.ok) {
                msg.innerText = "Password updated successfully!";
                msg.style.color = "green";
                msg.style.display = "block";
                togglePasswordEdit();
            } else {
                msg.innerText = await res.text();
                msg.style.color = "red";
                msg.style.display = "block";
            }
        } catch (err) {
            console.error(err);
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