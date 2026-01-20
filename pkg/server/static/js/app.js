let refreshInterval = 5;
let refreshTimer = null;
let autoRefresh = true;

async function fetchAllMetrics() {
    try {
        const [cpu, gpu, battery, disks, services] = await Promise.all([
            fetch('/api/v1/cpu?show-temp=true').then(r => r.json()),
            fetch('/api/v1/gpu').then(r => r.json()),
            fetch('/api/v1/battery').then(r => r.json()),
            fetch('/api/v1/disk').then(r => r.json()),
            fetch('/api/v1/services').then(r => r.json())
        ]);

        return { cpu, gpu, battery, disks, services };
    } catch (error) {
        console.error('Error fetching metrics:', error);
        return null;
    }
}

function updateDashboard(metrics) {
    if (!metrics) return;

    updateCPU(metrics.cpu);
    updateGPU(metrics.gpu);
    updateBattery(metrics.battery);
    updateDisks(metrics.disks);
    updateServices(metrics.services);
}

function updateCPU(cpuData) {
    if (cpuData.status !== 'success') return;

    const usage = cpuData.data.usagePercent || 0;
    const temp = cpuData.data.cpuTemperature || 0;

    document.getElementById('cpu-usage').textContent = usage.toFixed(1) + '%';
    document.getElementById('cpu-bar').style.width = usage + '%';
    document.getElementById('cpu-temp').textContent = temp.toFixed(1) + '°C';

    updateProgressBarColor('cpu-bar', usage);
}

function updateGPU(gpuData) {
    if (gpuData.status !== 'success') return;

    const gpu = gpuData.data[0];
    if (!gpu) return;

    const usage = gpu.gpuUtilPercent || 0;
    const temp = gpu.temperature || 0;

    document.getElementById('gpu-usage').textContent = usage.toFixed(1) + '%';
    document.getElementById('gpu-bar').style.width = usage + '%';
    document.getElementById('gpu-temp').textContent = temp.toFixed(1) + '°C';

    updateProgressBarColor('gpu-bar', usage);
}

function updateBattery(batteryData) {
    if (batteryData.status !== 'success') return;

    const battery = batteryData.data;
    if (!battery) return;

    const level = battery.capacity || 0;
    const status = battery.status || 'Unknown';

    document.getElementById('battery-level').textContent = level.toFixed(1) + '%';
    document.getElementById('battery-bar').style.width = level + '%';
    document.getElementById('battery-status').textContent = status;

    updateProgressBarColor('battery-bar', level);
}

function updateDisks(disksData) {
    if (disksData.status !== 'success') return;

    const disksList = document.getElementById('disks-list');
    disksList.innerHTML = '';

    if (!disksData.data || !Array.isArray(disksData.data)) {
        disksList.innerHTML = '<p>No disks found</p>';
        return;
    }

    disksData.data.forEach(disk => {
        const usage = disk.usagePercent || 0;
        const size = formatBytes(disk.size);
        const used = formatBytes(disk.used);
        const statusClass = getStatusClass(usage);

        const diskHtml = `
            <div class="disk-item ${statusClass}">
                <div class="metric">
                    <span class="label">${disk.name} (${disk.mountPoint || 'Not mounted'})</span>
                    <span class="value">${used} / ${size}</span>
                </div>
                <div class="progress-bar">
                    <div class="progress-fill" style="width: ${usage}%; background: ${getProgressColor(usage)}"></div>
                </div>
            </div>
        `;
        disksList.innerHTML += diskHtml;
    });
}

function updateServices(servicesData) {
    if (servicesData.status !== 'success') return;

    const data = servicesData.data;
    document.getElementById('services-running').textContent = data.running || 0;
    document.getElementById('services-stopped').textContent = data.stopped || 0;
    document.getElementById('services-failed').textContent = data.failed || 0;

    const servicesList = document.getElementById('services-list');
    servicesList.innerHTML = '';

    if (!data.services || !Array.isArray(data.services)) {
        servicesList.innerHTML = '<p>No services found</p>';
        return;
    }

    data.services.slice(0, 10).forEach(service => {
        const statusClass = service.status.toLowerCase();

        const serviceHtml = `
            <div class="service-item">
                <span class="service-name">${service.name}</span>
                <span class="service-status ${statusClass}">${service.status}</span>
            </div>
        `;
        servicesList.innerHTML += serviceHtml;
    });
}

function formatBytes(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function updateProgressBarColor(barId, percentage) {
    const bar = document.getElementById(barId);
    if (percentage >= 90) {
        bar.style.background = 'linear-gradient(90deg, #ff0000 0%, #ff4500 100%)';
    } else if (percentage >= 75) {
        bar.style.background = 'linear-gradient(90deg, #ffff00 0%, #ffa500 100%)';
    } else {
        bar.style.background = 'linear-gradient(90deg, #00ff00 0%, #7fff00 100%)';
    }
}

function getProgressColor(percentage) {
    if (percentage >= 90) return 'linear-gradient(90deg, #ff0000 0%, #ff4500 100%)';
    if (percentage >= 75) return 'linear-gradient(90deg, #ffff00 0%, #ffa500 100%)';
    return 'linear-gradient(90deg, #00ff00 0%, #7fff00 100%)';
}

function getStatusClass(percentage) {
    if (percentage >= 90) return 'critical';
    if (percentage >= 75) return 'warning';
    return '';
}

async function autoRefreshLoop() {
	if (!autoRefresh || refreshInterval === 0) return;

	const metrics = await fetchAllMetrics();
	if (metrics) {
		updateDashboard(metrics);
	}

	if (autoRefresh && refreshInterval > 0) {
		refreshTimer = setTimeout(autoRefreshLoop, refreshInterval * 1000);
	}
}

function startAutoRefresh() {
    if (refreshTimer) clearTimeout(refreshTimer);
    autoRefreshLoop();
}

function stopAutoRefresh() {
    if (refreshTimer) {
        clearTimeout(refreshTimer);
        refreshTimer = null;
    }
}

document.addEventListener('DOMContentLoaded', () => {
    fetchAllMetrics().then(metrics => {
        if (metrics) {
            updateDashboard(metrics);
        }
        startAutoRefresh();
    });

    document.getElementById('refresh-interval').addEventListener('change', (e) => {
        refreshInterval = parseInt(e.target.value);
        if (refreshInterval > 0 && autoRefresh) {
            startAutoRefresh();
        } else {
            stopAutoRefresh();
        }
    });

    document.getElementById('toggle-refresh').addEventListener('click', (e) => {
        autoRefresh = !autoRefresh;
        e.target.textContent = autoRefresh ? '⏸️ Pause' : '▶️ Resume';
        if (autoRefresh && refreshInterval > 0) {
            startAutoRefresh();
        } else {
            stopAutoRefresh();
        }
    });

    document.getElementById('refresh-now').addEventListener('click', () => {
        fetchAllMetrics().then(metrics => {
            if (metrics) updateDashboard(metrics);
        });
    });
});
