// Sample member data
const members = [
    { name: "Kwame Asante", role: "President", initial: "KA" },
    { name: "Kofi Mensah", role: "Vice President", initial: "KM" },
    { name: "Yaw Owusu", role: "Secretary", initial: "YO" },
    { name: "Kweku Osei", role: "Treasurer", initial: "KO" },
    { name: "Ama Adjei", role: "Financial Secretary", initial: "AA" },
    { name: "Akosua Boateng", role: "Organizer", initial: "AB" },
    { name: "Kwadwo Nyong", role: "Public Relations", initial: "KN" },
    { name: "Nana Okoye", role: "Member", initial: "NO" },
    { name: "Kojo Amponsah", role: "Member", initial: "KA" },
    { name: "Emmanuel Tetteh", role: "Member", initial: "ET" },
    { name: "Samuel Nkrumah", role: "Member", initial: "SN" },
    { name: "Richard Appiah", role: "Member", initial: "RA" },
];

// Navigation functionality
const navLinks = document.querySelectorAll(".nav-link");
const sections = document.querySelectorAll(".section");
const hamburger = document.querySelector(".hamburger");
const navMenu = document.querySelector(".nav-menu");

// Handle navigation
navLinks.forEach((link) => {
    link.addEventListener("click", (e) => {
        e.preventDefault();
        const targetSection = link.getAttribute("data-section");

        // Hide all sections
        sections.forEach((section) => {
            section.classList.remove("active");
        });

        // Show target section
        document.getElementById(targetSection).classList.add("active");

        // Close mobile menu
        navMenu.classList.remove("active");

        // Smooth scroll to top
        window.scrollTo({ top: 0, behavior: "smooth" });
    });
});

// CTA button functionality
document.querySelector(".cta-button").addEventListener("click", (e) => {
    e.preventDefault();
    const targetSection = e.target.getAttribute("data-section");

    sections.forEach((section) => {
        section.classList.remove("active");
    });

    document.getElementById(targetSection).classList.add("active");
    window.scrollTo({ top: 0, behavior: "smooth" });
});

// Mobile menu toggle
hamburger.addEventListener("click", () => {
    navMenu.classList.toggle("active");
});

// Populate members
function populateMembers() {
    const membersGrid = document.getElementById("membersGrid");
    membersGrid.innerHTML = "";

    members.forEach((member) => {
        const memberCard = document.createElement("div");
        memberCard.className = "member-card";
        memberCard.innerHTML = `
                    <div class="member-avatar">
                        ${member.initial}
                    </div>
                    <div class="member-info">
                        <div class="member-name">${member.name}</div>
                        <div class="member-role">${member.role}</div>
                    </div>
                `;
        membersGrid.appendChild(memberCard);
    });
}

// Form handling
function showNotification(message, type = "success") {
    const notification = document.getElementById("notification");
    notification.textContent = message;
    notification.style.background =
        type === "success" ? "var(--success)" : "#ef4444";
    notification.classList.add("show");

    setTimeout(() => {
        notification.classList.remove("show");
    }, 4000);
}

// Join form handling
document.getElementById("joinForm").addEventListener("submit", (e) => {
    e.preventDefault();
    const formData = new FormData(e.target);
    const data = Object.fromEntries(formData.entries());

    // Validate Guggisberg House requirement
    if (data.house !== "guggisberg") {
        showNotification(
            "Currently, membership is exclusive to Guggisberg House alumni.",
            "error"
        );
        return;
    }

    // Simulate form submission
    showNotification(
        "Application submitted successfully! We'll review and get back to you soon."
    );
    e.target.reset();
});

// Contact form handling
document.getElementById("contactForm").addEventListener("submit", (e) => {
    e.preventDefault();
    showNotification("Message sent successfully! We'll get back to you soon.");
    e.target.reset();
});

// Counter animation for stats
function animateCounter(element, start, end, duration) {
    let startTimestamp = null;
    const step = (timestamp) => {
        if (!startTimestamp) startTimestamp = timestamp;
        const progress = Math.min((timestamp - startTimestamp) / duration, 1);
        element.innerHTML = Math.floor(progress * (end - start) + start);
        if (progress < 1) {
            window.requestAnimationFrame(step);
        }
    };
    window.requestAnimationFrame(step);
}

// Intersection Observer for animations
const observerOptions = {
    threshold: 0.1,
    rootMargin: "0px 0px -50px 0px",
};

const observer = new IntersectionObserver((entries) => {
    entries.forEach((entry) => {
        if (entry.isIntersecting) {
            // Animate stats when they come into view
            if (entry.target.id === "memberCount") {
                animateCounter(entry.target, 0, 45, 2000);
            }

            // Add animation classes to cards
            if (
                entry.target.classList.contains("about-card") ||
                entry.target.classList.contains("member-card") ||
                entry.target.classList.contains("achievement-item")
            ) {
                entry.target.style.opacity = "0";
                entry.target.style.transform = "translateY(30px)";
                entry.target.style.transition = "all 0.6s ease";

                setTimeout(() => {
                    entry.target.style.opacity = "1";
                    entry.target.style.transform = "translateY(0)";
                }, 100);
            }
        }
    });
}, observerOptions);

// Initialize
document.addEventListener("DOMContentLoaded", () => {
    populateMembers();

    // Observe elements for animations
    document
        .querySelectorAll(
            ".about-card, .member-card, .achievement-item, #memberCount"
        )
        .forEach((el) => {
            observer.observe(el);
        });

    // Add some interactive features
    addInteractiveFeatures();
});

// Additional interactive features
function addInteractiveFeatures() {
    // Add ripple effect to buttons
    document.querySelectorAll(".cta-button, .form-button").forEach((button) => {
        button.addEventListener("click", function (e) {
            const ripple = document.createElement("span");
            const rect = this.getBoundingClientRect();
            const size = Math.max(rect.width, rect.height);
            const x = e.clientX - rect.left - size / 2;
            const y = e.clientY - rect.top - size / 2;

            ripple.style.width = ripple.style.height = size + "px";
            ripple.style.left = x + "px";
            ripple.style.top = y + "px";
            ripple.classList.add("ripple");

            this.appendChild(ripple);

            setTimeout(() => {
                ripple.remove();
            }, 600);
        });
    });

    // Add hover effects to member cards
    document.querySelectorAll(".member-card").forEach((card) => {
        card.addEventListener("mouseenter", function () {
            this.style.transform = "translateY(-10px) scale(1.02)";
        });

        card.addEventListener("mouseleave", function () {
            this.style.transform = "translateY(0) scale(1)";
        });
    });

    // Add typing effect to hero text (optional enhancement)
    const heroTitle = document.querySelector(".hero h1");
    const heroText = document.querySelector(".hero p");

    // Add search functionality for members (bonus feature)
    addMemberSearch();
}

// Member search functionality
function addMemberSearch() {
    const membersSection = document.getElementById("members");
    const membersGrid = document.getElementById("membersGrid");

    // Create search input
    const searchContainer = document.createElement("div");
    searchContainer.style.textAlign = "center";
    searchContainer.style.marginBottom = "2rem";

    const searchInput = document.createElement("input");
    searchInput.type = "text";
    searchInput.placeholder = "Search members...";
    searchInput.style.padding = "1rem";
    searchInput.style.borderRadius = "25px";
    searchInput.style.border = "2px solid #e2e8f0";
    searchInput.style.width = "300px";
    searchInput.style.maxWidth = "90%";
    searchInput.style.fontSize = "1rem";

    searchContainer.appendChild(searchInput);

    // Insert search before members grid
    membersSection
        .querySelector(".section")
        .insertBefore(searchContainer, membersGrid);

    // Search functionality
    searchInput.addEventListener("input", (e) => {
        const searchTerm = e.target.value.toLowerCase();
        const memberCards = document.querySelectorAll(".member-card");

        memberCards.forEach((card) => {
            const name = card
                .querySelector(".member-name")
                .textContent.toLowerCase();
            const role = card
                .querySelector(".member-role")
                .textContent.toLowerCase();

            if (name.includes(searchTerm) || role.includes(searchTerm)) {
                card.style.display = "block";
                card.style.animation = "slideInUp 0.3s ease";
            } else {
                card.style.display = "none";
            }
        });
    });
}

// Add smooth scrolling for internal navigation
function smoothScroll(target) {
    const element = document.getElementById(target);
    if (element) {
        element.scrollIntoView({
            behavior: "smooth",
            block: "start",
        });
    }
}

// Add dynamic year updates
function updateCurrentYear() {
    const currentYear = new Date().getFullYear();
    const yearsSinceGraduation = currentYear - 2018;

    // Update stats dynamically
    document.querySelector(".stat-item:last-child .stat-number").textContent =
        yearsSinceGraduation;
}

// Theme toggle functionality (bonus feature)
function addThemeToggle() {
    const themeToggle = document.createElement("button");
    themeToggle.innerHTML = '<i class="fas fa-moon"></i>';
    themeToggle.style.position = "fixed";
    themeToggle.style.bottom = "30px";
    themeToggle.style.right = "30px";
    themeToggle.style.width = "60px";
    themeToggle.style.height = "60px";
    themeToggle.style.borderRadius = "50%";
    themeToggle.style.border = "none";
    themeToggle.style.background = "var(--primary-blue)";
    themeToggle.style.color = "white";
    themeToggle.style.fontSize = "1.5rem";
    themeToggle.style.cursor = "pointer";
    themeToggle.style.boxShadow = "0 10px 30px rgba(0,0,0,0.2)";
    themeToggle.style.transition = "all 0.3s ease";
    themeToggle.style.zIndex = "1000";

    document.body.appendChild(themeToggle);

    let isDark = false;
    themeToggle.addEventListener("click", () => {
        if (!isDark) {
            // Switch to dark theme
            document.documentElement.style.setProperty(
                "--primary-blue",
                "#1e293b"
            );
            document.documentElement.style.setProperty(
                "--secondary-blue",
                "#334155"
            );
            document.documentElement.style.setProperty("--white", "#0f172a");
            document.documentElement.style.setProperty(
                "--light-gray",
                "#1e293b"
            );
            document.documentElement.style.setProperty(
                "--dark-gray",
                "#e2e8f0"
            );
            themeToggle.innerHTML = '<i class="fas fa-sun"></i>';
            document.body.style.background = "#0f172a";
            document.body.style.color = "#e2e8f0";
        } else {
            // Switch back to light theme
            document.documentElement.style.setProperty(
                "--primary-blue",
                "#1e40af"
            );
            document.documentElement.style.setProperty(
                "--secondary-blue",
                "#3b82f6"
            );
            document.documentElement.style.setProperty("--white", "#ffffff");
            document.documentElement.style.setProperty(
                "--light-gray",
                "#f8fafc"
            );
            document.documentElement.style.setProperty(
                "--dark-gray",
                "#334155"
            );
            themeToggle.innerHTML = '<i class="fas fa-moon"></i>';
            document.body.style.background = "";
            document.body.style.color = "";
        }
        isDark = !isDark;
    });
}

// Initialize all features
setTimeout(() => {
    updateCurrentYear();
    addThemeToggle();
}, 1000);

// Add CSS for ripple effect
const style = document.createElement("style");
style.textContent = `
            .ripple {
                position: absolute;
                border-radius: 50%;
                background-color: rgba(255, 255, 255, 0.6);
                transform: scale(0);
                animation: ripple 0.6s linear;
                pointer-events: none;
            }
            
            @keyframes ripple {
                to {
                    transform: scale(4);
                    opacity: 0;
                }
            }
            
            button {
                position: relative;
                overflow: hidden;
            }
        `;
document.head.appendChild(style);
