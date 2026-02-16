class SalesTrackerApp {
    constructor() {
        this.apiUrl = window.location.origin;
        this.currentPage = 1;
        this.limit = 25;
        this.allItems = [];
        this.filteredItems = [];
        this.totalItems = 0;
        this.analyticData = null;
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.loadItems();
        this.setupDateTimePickers();
    }

    setupDateTimePickers() {
        const now = new Date();
        const setDateTime = (elementId) => {
            const year = now.getFullYear();
            const month = String(now.getMonth() + 1).padStart(2, '0');
            const day = String(now.getDate()).padStart(2, '0');
            const hours = String(now.getHours()).padStart(2, '0');
            const minutes = String(now.getMinutes()).padStart(2, '0');
            document.getElementById(elementId).value = `${year}-${month}-${day}T${hours}:${minutes}`;
        };
        setDateTime('item-date');
        setDateTime('analytics-from');
        setDateTime('analytics-to');
        
        const yesterday = new Date();
        yesterday.setDate(yesterday.getDate() - 7);
        document.getElementById('filter-date-from').value = yesterday.toISOString().split('T')[0];
        document.getElementById('filter-date-to').value = new Date().toISOString().split('T')[0];
    }

    setupEventListeners() {
        document.querySelectorAll('.tab-btn').forEach(btn => {
            btn.addEventListener('click', () => this.switchTab(btn.dataset.tab));
        });
        
        document.getElementById('apply-filters').addEventListener('click', () => this.applyFilters());
        document.getElementById('reset-filters').addEventListener('click', () => this.resetFilters());
        
        document.getElementById('prev-page').addEventListener('click', () => {
            if (this.currentPage > 1) {
                this.currentPage--;
                this.renderItems();
                this.updatePagination();
            }
        });
        
        document.getElementById('next-page').addEventListener('click', () => {
            this.currentPage++;
            this.renderItems();
            this.updatePagination();
        });
        
        document.getElementById('add-item-form').addEventListener('submit', e => {
            e.preventDefault();
            this.createItem();
        });
        
        document.getElementById('clear-form').addEventListener('click', () => {
            document.getElementById('add-item-form').reset();
            this.setupDateTimePickers();
        });
        
        document.querySelector('.close').addEventListener('click', () => this.closeModal());
        document.getElementById('delete-item-btn').addEventListener('click', () => this.deleteItem());
        document.getElementById('edit-item-form').addEventListener('submit', e => {
            e.preventDefault();
            this.updateItem();
        });
        
        window.addEventListener('click', e => {
            if (e.target === document.getElementById('modal')) this.closeModal();
        });
        
        document.getElementById('get-analytics').addEventListener('click', () => this.getAnalytics());
        document.getElementById('export-analytics-csv').addEventListener('click', () => this.exportToCSV());
    }

    switchTab(tabName) {
        document.querySelectorAll('.tab-content').forEach(tab => tab.classList.remove('active'));
        document.querySelectorAll('.tab-btn').forEach(btn => btn.classList.remove('active'));
        document.getElementById(`${tabName}-tab`).classList.add('active');
        document.querySelector(`[data-tab="${tabName}"]`).classList.add('active');
    }

    async loadItems() {
        try {
            const response = await fetch(`${this.apiUrl}/items`);
            if (!response.ok) {
                throw new Error(`Ошибка сервера: ${response.status}`);
            }
            const data = await response.json();
            if (!data.items || !Array.isArray(data.items)) {
                throw new Error('Некорректные данные от сервера');
            }
            this.allItems = data.items;
            this.filteredItems = [...this.allItems];
            this.totalItems = this.filteredItems.length;
            this.currentPage = 1;
            this.renderItems();
            this.updatePagination();
        } catch (error) {
            console.error('Ошибка загрузки записей:', error);
            this.showErrorMessage('Не удалось загрузить записи. Проверьте подключение к серверу.');
            document.getElementById('items-body').innerHTML =
                '<tr><td colspan="7" class="no-data">Ошибка загрузки данных</td></tr>';
        }
    }

    applyFilters() {
        const type = document.getElementById('filter-type').value;
        const category = document.getElementById('filter-category').value.trim();
        const dateFrom = document.getElementById('filter-date-from').value;
        const dateTo = document.getElementById('filter-date-to').value;
        
        this.filteredItems = this.allItems.filter(item => {
            if (type && item.type !== type) {
                return false;
            }
            if (category && item.category && !item.category.toLowerCase().includes(category.toLowerCase())) {
                return false;
            }
            if (dateFrom || dateTo) {
                const itemDate = new Date(item.date);
                if (dateFrom) {
                    const fromDate = new Date(dateFrom);
                    fromDate.setHours(0, 0, 0, 0);
                    if (itemDate < fromDate) {
                        return false;
                    }
                }
                if (dateTo) {
                    const toDate = new Date(dateTo);
                    toDate.setHours(23, 59, 59, 999);
                    if (itemDate > toDate) {
                        return false;
                    }
                }
            }
            return true;
        });
        
        this.totalItems = this.filteredItems.length;
        this.currentPage = 1;
        this.renderItems();
        this.updatePagination();
    }

    resetFilters() {
        document.getElementById('filter-type').value = '';
        document.getElementById('filter-category').value = '';
        const today = new Date();
        const yesterday = new Date();
        yesterday.setDate(yesterday.getDate() - 7);
        document.getElementById('filter-date-from').value = yesterday.toISOString().split('T')[0];
        document.getElementById('filter-date-to').value = today.toISOString().split('T')[0];
        this.filteredItems = [...this.allItems];
        this.totalItems = this.filteredItems.length;
        this.currentPage = 1;
        this.renderItems();
        this.updatePagination();
    }

    renderItems() {
        const tbody = document.getElementById('items-body');
        const startIndex = (this.currentPage - 1) * this.limit;
        const endIndex = startIndex + this.limit;
        const itemsToShow = this.filteredItems.slice(startIndex, endIndex);
        
        if (itemsToShow.length === 0) {
            tbody.innerHTML = '<tr><td colspan="7" class="no-data">Записей не найдено</td></tr>';
            return;
        }
        
        tbody.innerHTML = itemsToShow.map(item => `
            <tr>
                <td>${item.id}</td>
                <td><span class="${item.type}">${this.getTypeLabel(item.type)}</span></td>
                <td><span class="${item.type}">${this.formatCurrency(item.amount)}</span></td>
                <td>${this.formatDate(item.date)}</td>
                <td>${item.category || '-'}</td>
                <td>${item.description || '-'}</td>
                <td class="actions">
                    <button class="action-btn edit" onclick="app.openEditModal(${item.id})">✏️</button>
                    <button class="action-btn delete" onclick="app.confirmDelete(${item.id})">🗑️</button>
                </td>
            </tr>
        `).join('');
    }

    updatePagination() {
        document.getElementById('total-count').textContent = `Всего записей: ${this.totalItems}`;
        document.getElementById('page-info').textContent = `Страница ${this.currentPage}`;
        document.getElementById('prev-page').disabled = this.currentPage === 1;
        document.getElementById('next-page').disabled = this.currentPage * this.limit >= this.totalItems;
    }

    getTypeLabel(type) {
        return type === 'income' ? 'Доход' : 'Расход';
    }

    formatCurrency(amount) {
        if (amount === undefined || amount === null) return '0.00 ₽';
        return new Intl.NumberFormat('ru-RU', {
            style: 'currency',
            currency: 'RUB',
            minimumFractionDigits: 2
        }).format(amount);
    }

    formatDate(dateString) {
        if (!dateString) return '-';
        try {
            const date = new Date(dateString);
            if (isNaN(date)) return '-';
            return date.toLocaleDateString('ru-RU', {
                year: 'numeric',
                month: '2-digit',
                day: '2-digit',
                hour: '2-digit',
                minute: '2-digit'
            });
        } catch (e) {
            return '-';
        }
    }

    async createItem() {
        const formData = {
            type: document.getElementById('item-type').value,
            amount: parseFloat(document.getElementById('item-amount').value),
            date: this.convertToUTC(document.getElementById('item-date').value),
            category: document.getElementById('item-category').value.trim() || null,
            description: document.getElementById('item-description').value.trim() || null
        };
        
        if (!formData.type) {
            this.showErrorMessage('Выберите тип записи');
            return;
        }
        if (isNaN(formData.amount) || formData.amount <= 0) {
            this.showErrorMessage('Сумма должна быть положительным числом');
            return;
        }
        if (!formData.date) {
            this.showErrorMessage('Укажите дату');
            return;
        }
        
        try {
            const response = await fetch(`${this.apiUrl}/items`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(formData)
            });
            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(errorText || 'Ошибка создания записи');
            }
            this.showSuccessMessage('Запись успешно создана!');
            document.getElementById('add-item-form').reset();
            this.setupDateTimePickers();
            this.loadItems();
            this.switchTab('items');
        } catch (error) {
            this.showErrorMessage(`Ошибка создания записи: ${error.message}`);
        }
    }

    convertToUTC(localDateTimeStr) {
        if (!localDateTimeStr) return null;
        const withSeconds = localDateTimeStr.length === 16 ? `${localDateTimeStr}:00` : localDateTimeStr;
        const date = new Date(withSeconds);
        if (isNaN(date)) return null;
        return date.toISOString();
    }

    async openEditModal(id) {
        try {
            const response = await fetch(`${this.apiUrl}/items/${id}`);
            if (!response.ok) {
                throw new Error('Не удалось загрузить запись');
            }
            const item = await response.json();
            document.getElementById('edit-item-id').value = item.id;
            document.getElementById('edit-item-type').value = item.type;
            document.getElementById('edit-item-amount').value = item.amount;
            document.getElementById('edit-item-date').value = this.convertToLocal(item.date);
            document.getElementById('edit-item-category').value = item.category || '';
            document.getElementById('edit-item-description').value = item.description || '';
            document.getElementById('modal-title').textContent = `Редактировать запись #${item.id}`;
            document.getElementById('modal').style.display = 'block';
        } catch (error) {
            this.showErrorMessage(`Ошибка: ${error.message}`);
        }
    }

    convertToLocal(utcDateTimeStr) {
        if (!utcDateTimeStr) return '';
        try {
            const date = new Date(utcDateTimeStr);
            if (isNaN(date)) return '';
            const year = date.getFullYear();
            const month = String(date.getMonth() + 1).padStart(2, '0');
            const day = String(date.getDate()).padStart(2, '0');
            const hours = String(date.getHours()).padStart(2, '0');
            const minutes = String(date.getMinutes()).padStart(2, '0');
            return `${year}-${month}-${day}T${hours}:${minutes}`;
        } catch (e) {
            return '';
        }
    }

    async updateItem() {
        const id = document.getElementById('edit-item-id').value;
        const formData = {
            type: document.getElementById('edit-item-type').value,
            amount: parseFloat(document.getElementById('edit-item-amount').value),
            date: this.convertToUTC(document.getElementById('edit-item-date').value),
            category: document.getElementById('edit-item-category').value.trim() || null,
            description: document.getElementById('edit-item-description').value.trim() || null
        };
        
        if (!formData.type) {
            this.showErrorMessage('Выберите тип записи');
            return;
        }
        if (isNaN(formData.amount) || formData.amount <= 0) {
            this.showErrorMessage('Сумма должна быть положительным числом');
            return;
        }
        if (!formData.date) {
            this.showErrorMessage('Укажите дату');
            return;
        }
        
        try {
            const response = await fetch(`${this.apiUrl}/items/${id}`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(formData)
            });
            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(errorText || 'Ошибка обновления записи');
            }
            this.showSuccessMessage('Запись успешно обновлена!');
            this.closeModal();
            this.loadItems();
        } catch (error) {
            this.showErrorMessage(`Ошибка обновления записи: ${error.message}`);
        }
    }

    async confirmDelete(id) {
        if (!confirm('Вы уверены, что хотите удалить эту запись?')) return;
        try {
            const response = await fetch(`${this.apiUrl}/items/${id}`, {
                method: 'DELETE'
            });
            if (!response.ok) {
                throw new Error('Ошибка удаления записи');
            }
            this.showSuccessMessage('Запись успешно удалена!');
            this.loadItems();
        } catch (error) {
            this.showErrorMessage(`Ошибка удаления: ${error.message}`);
        }
    }

    closeModal() {
        document.getElementById('modal').style.display = 'none';
        document.getElementById('edit-item-form').reset();
    }

    async getAnalytics() {
        const fromInput = document.getElementById('analytics-from').value;
        const toInput = document.getElementById('analytics-to').value;
        
        if (!fromInput || !toInput) {
            this.showErrorMessage('Укажите обе даты для аналитики');
            return;
        }
        
        const from = this.convertToUTC(fromInput);
        const to = this.convertToUTC(toInput);
        
        if (!from || !to) {
            this.showErrorMessage('Некорректный формат даты');
            return;
        }
        
        try {
            const response = await fetch(`${this.apiUrl}/analytics?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`);
            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(errorText || 'Ошибка получения аналитики');
            }
            const data = await response.json();
            this.renderAnalytics(data);
        } catch (error) {
            this.showErrorMessage(`Ошибка получения аналитики: ${error.message}`);
            this.renderAnalytics({
                income: { sum: 0, avg: 0, count: 0, median: 0, percent90: 0 },
                expense: { sum: 0, avg: 0, count: 0, median: 0, percent90: 0 },
                details: []
            });
        }
    }

    renderAnalytics(data) {
        this.analyticData = data;
        
        // Доходы
        document.getElementById('analytics-income-sum').textContent = 
            this.formatCurrency(data.income?.sum || 0);
        document.getElementById('analytics-income-avg').textContent = 
            this.formatCurrency(data.income?.avg || 0);
        document.getElementById('analytics-income-count').textContent = 
            data.income?.count || 0;
        document.getElementById('analytics-income-median').textContent = 
            this.formatCurrency(data.income?.median || 0);
        document.getElementById('analytics-income-percent90').textContent = 
            this.formatCurrency(data.income?.percent90 || 0);
        
        // Расходы
        document.getElementById('analytics-expense-sum').textContent = 
            this.formatCurrency(data.expense?.sum || 0);
        document.getElementById('analytics-expense-avg').textContent = 
            this.formatCurrency(data.expense?.avg || 0);
        document.getElementById('analytics-expense-count').textContent = 
            data.expense?.count || 0;
        document.getElementById('analytics-expense-median').textContent = 
            this.formatCurrency(data.expense?.median || 0);
        document.getElementById('analytics-expense-percent90').textContent = 
            this.formatCurrency(data.expense?.percent90 || 0);
        
        this.renderAnalyticsDetails();
    }

    renderAnalyticsDetails() {
        const tbody = document.getElementById('analytics-body');
        if (!this.analyticData || !this.analyticData.details || this.analyticData.details.length === 0) {
            tbody.innerHTML = '<tr><td colspan="5" class="no-data">Нет записей в выбранном периоде</td></tr>';
            return;
        }
        
        tbody.innerHTML = this.analyticData.details.map(item => `
            <tr>
                <td>${item.id}</td>
                <td><span class="${item.type}">${this.getTypeLabel(item.type)}</span></td>
                <td><span class="${item.type}">${this.formatCurrency(item.amount)}</span></td>
                <td>${this.formatDate(item.date)}</td>
                <td>${item.category || '-'}</td>
            </tr>
        `).join('');
    }

    async exportToCSV() {
        try {
            const fromInput = document.getElementById('analytics-from').value;
            const toInput = document.getElementById('analytics-to').value;
            
            let exportUrl = `${this.apiUrl}/items/export`;
            
            // Если есть период из аналитики, передаём его
            if (fromInput && toInput) {
                const from = this.convertToUTC(fromInput);
                const to = this.convertToUTC(toInput);
                exportUrl += `?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`;
            }
            
            const link = document.createElement('a');
            link.href = exportUrl;
            link.download = `sales_tracker_${new Date().toISOString().slice(0, 10)}.csv`;
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);
            this.showSuccessMessage('Отчёт успешно экспортирован в CSV');
        } catch (error) {
            this.showErrorMessage(`Ошибка экспорта: ${error.message}`);
        }
    }

    showErrorMessage(message) {
        alert(`❌ Ошибка: ${message}`);
    }

    showSuccessMessage(message) {
        alert(`✅ Успех: ${message}`);
    }
}

document.addEventListener('DOMContentLoaded', () => {
    window.app = new SalesTrackerApp();
});