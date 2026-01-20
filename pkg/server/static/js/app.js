let refreshInterval = 5;
let refreshTimer = null;
let autoRefresh = true;

async function fetchAllMetrics() {
    try {
        const [cpu, gpu, battery, disks, services] = await Promise.all([
            fetch('/api/v1/cpu?show-cores=true&show-flags=true&show-temp=true').then(r => r.json()),
            fetch('/api/v1/gpu?show-connectors=true&show-modes=true').then(r => r.json()),
            fetch('/api/v1/battery').then(r => r.json()),
            fetch('/api/v1/disk?all=true&io-stats=true&inode-stats=true').then(r => r.json()),
            fetch('/api/v1/services').then(r => r.json())
        ]);

        return { cpu, gpu, battery, disks, services };
    } catch (error) {
        console.error('Error fetching metrics:', error);
        return null;
    }
}

function updateDashboard(metrics) {
    if (!metrics) {
        console.error('No metrics data received');
        return;
    }

    updateCPU(metrics.cpu);
    updateGPU(metrics.gpu);
    updateBattery(metrics.battery);
    updateDisks(metrics.disks);
    updateServices(metrics.services);
    updateLastUpdateTime();
}

function updateLastUpdateTime() {
    const now = new Date();
    const formatted = now.toISOString().replace('T', ' ').substring(0, 19);
    document.getElementById('last-update').textContent = `Last Update: ${formatted}`;
}

function updateCPU(cpuData) {
    if (cpuData.status !== 'success') {
        document.getElementById('cpu-content').innerHTML = '<p>Error loading CPU information</p>';
        return;
    }

    const cpu = cpuData.data;
    let html = '<table>';

    html += `<tr><td>Cores</td><td>${cpu.physicalCores} / ${cpu.logicalCores} (physical / logical)</td></tr>`;
    html += `<tr><td>Load Average</td><td>1m: ${cpu.loadAvg1.toFixed(2)} | 5m: ${cpu.loadAvg5.toFixed(2)} | 15m: ${cpu.loadAvg15.toFixed(2)}</td></tr>`;
    html += `<tr><td>Frequency</td><td>${formatFrequency(cpu.minFrequency)} / ${formatFrequency(cpu.maxFrequency)} / ${formatFrequency(cpu.currentFrequency)} (min / max / current)</td></tr>`;
    html += `<tr><td>Temperature</td><td>${cpu.cpuTemperature.toFixed(1)}°C 🌡️</td></tr>`;
    html += `<tr><td>Processes</td><td>${cpu.runningProcesses} running / ${cpu.totalProcesses} total</td></tr>`;

    html += '</table>';

    if (cpu.cores && cpu.cores.length > 0) {
        html += '<h3>🔧 Core Usage</h3>';
        html += '<table><thead><tr><th>Core</th><th>Physical ID</th><th>Frequency</th><th>Usage</th></tr></thead><tbody>';
        cpu.cores.forEach(core => {
            html += `<tr>
                <td>${core.coreId}</td>
                <td>${core.physicalId}</td>
                <td>${formatFrequency(core.frequency)}</td>
                <td>${core.usagePercent.toFixed(1)}% 📊</td>
            </tr>`;
        });
        html += '</tbody></table>';
    }

    html += '<h3>📋 Hardware Information</h3>';
    html += '<table>';
    html += `<tr><td>Model</td><td>${escapeHtml(cpu.modelName)}</td></tr>`;
    html += `<tr><td>Vendor</td><td>${escapeHtml(cpu.vendorId)}</td></tr>`;
    html += `<tr><td>Architecture</td><td>${escapeHtml(cpu.architecture)} 🖥️</td></tr>`;
    html += `<tr><td>CPU Family</td><td>${cpu.cpuFamily} / Model: ${cpu.model} / Stepping: ${cpu.stepping}</td></tr>`;
    html += '</table>';

    html += '<h3>💾 Cache Sizes</h3>';
    html += '<table>';
    html += `<tr><td>L1 Data</td><td>${cpu.cacheL1d} KB</td></tr>`;
    html += `<tr><td>L1 Instruction</td><td>${cpu.cacheL1i} KB</td></tr>`;
    html += `<tr><td>L2 Cache</td><td>${cpu.cacheL2} KB</td></tr>`;
    html += `<tr><td>L3 Cache</td><td>${cpu.cacheL3} KB (${(cpu.cacheL3 / 1024).toFixed(2)} MB)</td></tr>`;
    html += '</table>';

    if (cpu.flags && cpu.flags.length > 0) {
        html += '<h3>🏷️ CPU Flags</h3>';
        html += `<p>${cpu.flags.join(' ')}</p>`;
    }

    document.getElementById('cpu-content').innerHTML = html;
}

function updateGPU(gpuData) {
    if (gpuData.status !== 'success') {
        document.getElementById('gpu-content').innerHTML = '<p>Error loading GPU information</p>';
        return;
    }

    const gpus = gpuData.data;
    if (!gpus || gpus.length === 0) {
        document.getElementById('gpu-content').innerHTML = '<p>No GPUs detected</p>';
        return;
    }

    let html = '';

    gpus.forEach((gpu, index) => {
        html += `<h3>GPU ${index}: ${escapeHtml(gpu.cardName)}</h3>`;

        html += '<table>';
        html += `<tr><td>Status</td><td>${gpu.enabled ? '✅ Enabled' : '❌ Disabled'} | ${gpu.gpuType}</td></tr>`;
        html += `<tr><td>Driver</td><td>${escapeHtml(gpu.driver)} | Vendor: ${escapeHtml(gpu.vendor)}</td></tr>`;
        html += `<tr><td>Bus</td><td>${escapeHtml(gpu.busId)} | PCIe Gen: ${gpu.pcieGen} | Width: ${gpu.busWidth}</td></tr>`;

        html += `<tr><td>Temperature</td><td>${gpu.temperature.toFixed(1)}°C 🌡️ (Critical: ${gpu.temperatureCrit.toFixed(1)}°C)</td></tr>`;

        html += `<tr><td>GPU Usage</td><td>${gpu.gpuUtilPercent.toFixed(1)}% 📊 | Memory: ${gpu.memoryUtilPercent.toFixed(1)}%</td></tr>`;

        if (gpu.powerUsage !== undefined) {
            html += `<tr><td>Power</td><td>${gpu.powerUsage / 1000}W / Limit: ${gpu.powerLimit / 1000}W ⚡</td></tr>`;
        }

        if (gpu.fanSpeed !== undefined || gpu.fanSpeedPercent !== undefined) {
            const fanInfo = gpu.fanSpeed ? `${gpu.fanSpeed} RPM` : `${gpu.fanSpeedPercent.toFixed(1)}%`;
            html += `<tr><td>Fan Speed</td><td>💨 ${fanInfo}</td></tr>`;
        }

        html += '</table>';

        html += '<h3>💾 Memory</h3>';
        html += '<table>';
        html += `<tr><td>VRAM Total</td><td>${formatBytes(gpu.vramSize)}</td></tr>`;
        html += `<tr><td>VRAM Used</td><td>${formatBytes(gpu.vramUsed)} 📊</td></tr>`;
        html += `<tr><td>VRAM Free</td><td>${formatBytes(gpu.vramFree)}</td></tr>`;
        html += `<tr><td>GART Size</td><td>${formatBytes(gpu.gartSize)}</td></tr>`;
        html += '</table>';

        html += '<h3>⚡ Clocks</h3>';
        html += '<table>';
        html += `<tr><td>Core Clock</td><td>${formatFrequency(gpu.coreClock)} / Max: ${formatFrequency(gpu.maxCoreClock)}</td></tr>`;
        html += `<tr><td>Memory Clock</td><td>${formatFrequency(gpu.memoryClock)} / Max: ${formatFrequency(gpu.maxMemoryClock)}</td></tr>`;
        html += '</table>';

        if (gpu.connectors && gpu.connectors.length > 0) {
            html += '<h3>🔌 Connectors</h3>';
            html += '<table>';
            gpu.connectors.forEach(conn => {
                const enabled = gpu.enabledConnectors && gpu.enabledConnectors.includes(conn);
                html += `<tr><td>${escapeHtml(conn)}</td><td>${enabled ? '✅ Connected' : '⭕ Disconnected'}</td></tr>`;
            });
            html += '</table>';
        }

        if (gpu.modes && gpu.modes.length > 0) {
            html += '<h3>🖥️ Display Modes</h3>';
            html += '<table>';
            gpu.modes.slice(0, 10).forEach(mode => {
                html += `<tr><td>${escapeHtml(mode)}</td></tr>`;
            });
            if (gpu.modes.length > 10) {
                html += `<tr><td>... and ${gpu.modes.length - 10} more</td></tr>`;
            }
            html += '</table>';
        }
    });

    document.getElementById('gpu-content').innerHTML = html;
}

function updateDisks(disksData) {
    if (disksData.status !== 'success') {
        document.getElementById('disk-content').innerHTML = `<p>Error loading disk information: ${escapeHtml(disksData.error || 'Unknown error')}</p>`;
        return;
    }

    const disks = disksData.data.data || disksData.data;
    const ioStats = disksData.data.ioStats || {};
    const inodeInfo = disksData.data.inodeInfo || {};

    if (!disks || disks.length === 0) {
        document.getElementById('disk-content').innerHTML = '<p>No disks detected</p>';
        return;
    }

    let html = '<h3>💾 Disk Summary</h3>';
    html += '<table><thead><tr><th>Device</th><th>Mount</th><th>Size</th><th>Used</th><th>Available</th><th>Usage</th></tr></thead><tbody>';

    disks.forEach(disk => {
        const fsIcon = getFilesystemEmoji(disk.filesystemType);
        html += `<tr>
            <td>${disk.name} ${fsIcon}</td>
            <td>${escapeHtml(disk.mountPoint || 'Not mounted')}</td>
            <td>${formatBytes(disk.size)}</td>
            <td>${formatBytes(disk.used)}</td>
            <td>${formatBytes(disk.available)}</td>
            <td>${disk.usagePercent.toFixed(1)}% 📊</td>
        </tr>`;
    });

    html += '</tbody></table>';

    disks.forEach(disk => {
        html += `<h3>${disk.name} @ ${escapeHtml(disk.mountPoint || 'Not mounted')}</h3>`;

        html += '<table>';
        html += `<tr><td>Path</td><td>${escapeHtml(disk.path)}</td></tr>`;
        html += `<tr><td>Filesystem</td><td>${escapeHtml(disk.filesystemType)} 📁</td></tr>`;
        html += `<tr><td>Type</td><td>${disk.rotational ? '🔄 HDD' : '💿 SSD'} | ${disk.removable ? '📱 Removable' : '🔒 Fixed'}</td></tr>`;
        html += `<tr><td>UUID</td><td>${escapeHtml(disk.uuid)}</td></tr>`;
        html += `<tr><td>Label</td><td>${escapeHtml(disk.label || 'N/A')}</td></tr>`;
        html += `<tr><td>Block Size</td><td>${formatBytes(disk.blockSize)}</td></tr>`;

        if (disk.model) {
            html += `<tr><td>Model</td><td>${escapeHtml(disk.model)}</td></tr>`;
        }
        if (disk.serial) {
            html += `<tr><td>Serial</td><td>${escapeHtml(disk.serial)}</td></tr>`;
        }

        html += '</table>';

        if (ioStats && ioStats[disk.name]) {
            const io = ioStats[disk.name];
            html += '<h3>⚡ I/O Statistics</h3>';
            html += '<table>';
            html += `<tr><td>Reads</td><td>${io.readsCompleted.toLocaleString()} (${formatBytes(io.sectorsRead * 512)})</td></tr>`;
            html += `<tr><td>Writes</td><td>${io.writesCompleted.toLocaleString()} (${formatBytes(io.sectorsWritten * 512)})</td></tr>`;
            html += `<tr><td>Read Time</td><td>${io.readTimeMs.toLocaleString()} ms</td></tr>`;
            html += `<tr><td>Write Time</td><td>${io.writeTimeMs.toLocaleString()} ms</td></tr>`;
            html += `<tr><td>Total I/O Time</td><td>${io.ioTimeMs.toLocaleString()} ms</td></tr>`;
            html += '</table>';
        }

        if (inodeInfo && inodeInfo[disk.name]) {
            const inode = inodeInfo[disk.name];
            html += '<h3>📊 Inode Usage</h3>';
            html += '<table>';
            html += `<tr><td>Total</td><td>${inode.total.toLocaleString()}</td></tr>`;
            html += `<tr><td>Used</td><td>${inode.used.toLocaleString()}</td></tr>`;
            html += `<tr><td>Free</td><td>${inode.free.toLocaleString()}</td></tr>`;
            html += `<tr><td>Usage</td><td>${inode.usagePercent.toFixed(1)}% 📊</td></tr>`;
            html += '</table>';
        }
    });

    document.getElementById('disk-content').innerHTML = html;
}

function updateServices(servicesData) {
    if (servicesData.status !== 'success') {
        document.getElementById('services-content').innerHTML = `<p>Error loading services information: ${escapeHtml(servicesData.error || 'Unknown error')}</p>`;
        return;
    }

    const data = servicesData.data;

    let html = '<div class="summary-cards">';
    html += `<div class="card"><h3>Total Services</h3><div class="value">${data.totalServices}</div></div>`;
    html += `<div class="card"><h3>Running</h3><div class="value">${data.running} 🟢</div></div>`;
    html += `<div class="card"><h3>Stopped</h3><div class="value">${data.stopped} ⏸️</div></div>`;
    html += `<div class="card"><h3>Failed</h3><div class="value">${data.failed} ❌</div></div>`;
    html += `<div class="card"><h3>Listening Ports</h3><div class="value">${data.totalPorts}</div></div>`;
    html += '</div>';

    if (data.services && data.services.length > 0) {
        html += '<h3>📋 Services List</h3>';
        html += '<table><thead><tr><th>Name</th><th>Status</th><th>PID</th><th>Memory</th><th>CPU</th><th>User</th><th>Ports</th></tr></thead><tbody>';

        data.services.forEach(service => {
            const statusIcon = service.status === 'running' ? '🟢' :
                              service.status === 'stopped' ? '⏸️' : '❌';

            const pid = service.pid > 0 ? service.pid : '-';
            const memory = service.pid > 0 ? `${service.memoryMB.toFixed(1)} MB` : '-';
            const cpu = service.pid > 0 ? `${service.cpuPercent.toFixed(2)}%` : '-';

            let ports = '-';
            if (service.listeningPorts && service.listeningPorts.length > 0) {
                ports = service.listeningPorts.map(p => {
                    const protocolIcon = p.protocol.toLowerCase().includes('tcp') ? '🟢' : '🔵';
                    return `${protocolIcon}:${p.port}`;
                }).join('<br>');
            }

            html += `<tr>
                <td><strong>${escapeHtml(service.name)}</strong></td>
                <td>${statusIcon} ${escapeHtml(service.status)}</td>
                <td>${pid}</td>
                <td>${memory}</td>
                <td>${cpu}</td>
                <td>${escapeHtml(service.user)}</td>
                <td>${ports}</td>
            </tr>`;
        });

        html += '</tbody></table>';
    }

    document.getElementById('services-content').innerHTML = html;
}

function updateBattery(batteryData) {
    if (batteryData.status !== 'success') {
        document.getElementById('battery-content').innerHTML = `<p>Error loading battery information: ${escapeHtml(batteryData.error || 'No battery detected')}</p>`;
        return;
    }

    const battery = batteryData.data;
    if (!battery) {
        document.getElementById('battery-content').innerHTML = '<p>No battery detected</p>';
        return;
    }

    let html = '<div class="summary-cards">';
    html += `<div class="card"><h3>Capacity</h3><div class="value">${battery.capacity.toFixed(0)}% ${getBatteryEmoji(battery.capacity)}</div></div>`;

    if (battery.healthPercent !== undefined && battery.healthPercent !== 0) {
        html += `<div class="card"><h3>Health</h3><div class="value">${battery.healthPercent.toFixed(0)}% ${getHealthEmoji(battery.healthPercent)}</div></div>`;
    } else if (battery.health !== undefined) {
        html += `<div class="card"><h3>Health</h3><div class="value">${escapeHtml(battery.health)}</div></div>`;
    }
    html += '</div>';

    html += '<table>';
    html += `<tr><td>Status</td><td>${escapeHtml(battery.status)} ${getChargingEmoji(battery.status)}</td></tr>`;

    if (battery.technology) {
        html += `<tr><td>Technology</td><td>${escapeHtml(battery.technology)}</td></tr>`;
    }
    if (battery.manufacturer) {
        html += `<tr><td>Manufacturer</td><td>${escapeHtml(battery.manufacturer)}</td></tr>`;
    }
    if (battery.modelName) {
        html += `<tr><td>Model</td><td>${escapeHtml(battery.modelName)}</td></tr>`;
    }
    if (battery.serialNumber) {
        html += `<tr><td>Serial Number</td><td>${escapeHtml(battery.serialNumber)}</td></tr>`;
    }

    if (battery.voltage !== undefined) {
        html += `<tr><td>Voltage</td><td>${battery.voltage.toFixed(2)}V ⚡</td></tr>`;
    }
    if (battery.current !== undefined) {
        html += `<tr><td>Current</td><td>${battery.current.toFixed(2)}A</td></tr>`;
    }

    if (battery.currentFullCap !== undefined) {
        html += `<tr><td>Current Full Capacity</td><td>${battery.currentFullCap.toFixed(2)}Wh</td></tr>`;
    }
    if (battery.designCapacity !== undefined) {
        html += `<tr><td>Design Capacity</td><td>${battery.designCapacity.toFixed(2)}Wh</td></tr>`;
    }

    if (battery.cycleCount !== undefined) {
        html += `<tr><td>Cycle Count</td><td>${battery.cycleCount} 🔄</td></tr>`;
    }

    if (battery.health) {
        html += `<tr><td>Condition</td><td>${escapeHtml(battery.health)}</td></tr>`;
    }

    html += '</table>';

    document.getElementById('battery-content').innerHTML = html;
}

function formatBytes(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function formatFrequency(hz) {
    if (hz === 0) return '0 Hz';
    const mhz = hz / 1000000;
    if (mhz >= 1000) {
        return (mhz / 1000).toFixed(2) + ' GHz';
    }
    return mhz.toFixed(2) + ' MHz';
}

function getFilesystemEmoji(fsType) {
    const fs = fsType.toLowerCase();
    if (fs.includes('ext')) return '📁';
    if (fs.includes('xfs')) return '📂';
    if (fs.includes('btrfs')) return '🗄️';
    if (fs.includes('ntfs')) return '💿';
    if (fs.includes('fat') || fs.includes('exfat')) return '💾';
    if (fs.includes('swap')) return '🔄';
    if (fs.includes('tmpfs')) return '⚡';
    if (fs.includes('nfs') || fs.includes('smb')) return '🌐';
    if (fs.includes('zfs')) return '🌊';
    return '💿';
}

function getBatteryEmoji(capacity) {
    if (capacity >= 90) return '🔋';
    if (capacity >= 60) return '🔋';
    if (capacity >= 40) return '🪫';
    if (capacity >= 20) return '🪫';
    return '⚠️';
}

function getChargingEmoji(status) {
    const s = status.toLowerCase();
    if (s.includes('charg')) return '🔌';
    if (s.includes('discharg')) return '🔋';
    if (s.includes('full')) return '✅';
    return '';
}

function getHealthEmoji(healthPercent) {
    if (healthPercent >= 90) return '💚';
    if (healthPercent >= 75) return '💚';
    if (healthPercent >= 50) return '💛';
    if (healthPercent >= 25) return '🧡';
    return '❤️';
}

function escapeHtml(text) {
    if (!text) return '';
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
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
