// static/js/main.js
(function() {
    'use strict';

    // 页面加载完成后移除加载动画
    window.addEventListener('load', function() {
        const loader = document.getElementById('loader');
        if (loader) {
            loader.classList.add('hidden');
            setTimeout(() => loader.remove(), 500);
        }
    });

    // 导航栏滚动效果
    const header = document.querySelector('header');
    if (header) {
        let lastScroll = 0;
        window.addEventListener('scroll', function() {
            const currentScroll = window.pageYOffset;
            if (currentScroll > 50) {
                header.classList.add('shadow-md');
            } else {
                header.classList.remove('shadow-md');
            }
            lastScroll = currentScroll;
        }, { passive: true });
    }

    // 卡片进入视口时的动画
    const observerOptions = {
        threshold: 0.1,
        rootMargin: '0px 0px -50px 0px'
    };

    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.classList.add('animate-fade-in');
                observer.unobserve(entry.target);
            }
        });
    }, observerOptions);

    // 观察所有卡片
    document.querySelectorAll('.site-card').forEach(card => {
        observer.observe(card);
    });

    // HTMX 事件处理
    document.addEventListener('htmx:beforeRequest', function(evt) {
        // 显示加载状态
        const indicator = evt.detail.elt.querySelector('.htmx-indicator');
        if (indicator) {
            indicator.style.display = 'inline-block';
        }
    });

    document.addEventListener('htmx:afterRequest', function(evt) {
        // 隐藏加载状态
        const indicator = evt.detail.elt.querySelector('.htmx-indicator');
        if (indicator) {
            indicator.style.display = 'none';
        }
    });

    // 平滑滚动到顶部
    window.scrollToTop = function() {
        window.scrollTo({
            top: 0,
            behavior: 'smooth'
        });
    };

    // 复制到剪贴板
    window.copyToClipboard = function(text) {
        navigator.clipboard.writeText(text).then(() => {
            // 显示复制成功提示
            const toast = document.createElement('div');
            toast.className = 'fixed bottom-4 right-4 bg-green-500 text-white px-4 py-2 rounded-lg shadow-lg z-50 animate-fade-in';
            toast.textContent = '已复制到剪贴板';
            document.body.appendChild(toast);
            setTimeout(() => toast.remove(), 2000);
        });
    };

    // 图片加载失败处理
    document.querySelectorAll('img[onerror]').forEach(img => {
        img.addEventListener('error', function() {
            this.style.display = 'none';
            const fallback = this.nextElementSibling;
            if (fallback) {
                fallback.style.display = 'flex';
            }
        });
    });

})();
