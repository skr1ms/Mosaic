export const MainNav = [
    {
        icon: 'pe-7s-home',
        label: 'Главная панель',
        to: '/dashboard',
    },
];

export const AdminNav = [
    {
        icon: 'pe-7s-users',
        label: 'Управление партнерами',
        content: [
            {
                label: 'Список партнеров',
                to: '/partners/list',
            },
            {
                label: 'Добавить партнера',
                to: '/partners/create',
            },
            {
                label: 'Заблокированные партнеры',
                to: '/partners/blocked',
            },
        ],
            },
            {
        icon: 'pe-7s-ticket',
        label: 'Управление купонами',
        content: [
            {
                label: 'Создать купоны',
                to: '/coupons/create',
            },
            {
                label: 'Поиск и управление',
                to: '/coupons/manage',
            },
            {
                label: 'Активированные купоны',
                to: '/coupons/activated',
            },
            {
                label: 'Экспорт купонов',
                to: '/coupons/export',
            },
        ],
    },
    {
        icon: 'pe-7s-graph1',
        label: 'Статистика и аналитика',
        content: [
            {
                label: 'Общая статистика',
                to: '/analytics',
            },
            {
                label: 'Статистика по партнерам',
                to: '/analytics/partners',
            },
            {
                label: 'Активность системы',
                to: '/analytics/activity',
            },
        ],
    },
];

export const PartnerNav = [
    {
        icon: 'pe-7s-id',
        label: 'Партнерская панель',
        content: [
            {
                label: 'Мои купоны',
                to: '/partner/coupons',
            },
            {
                label: 'Моя статистика',
                to: '/partner/stats',
            },
            {
                label: 'Настройки профиля',
                to: '/partner/profile',
            },
        ],
    },
];

export const SystemNav = [
    {
        icon: 'pe-7s-config',
        label: 'Системные операции',
        content: [
            {
                label: 'Логи системы',
                to: '/system/logs',
    },
    {
                label: 'Резервные копии',
                to: '/system/backups',
    },
    {
                label: 'Настройки системы',
                to: '/system/settings',
            },
        ],
    },
];

// Оставляем для возможного будущего расширения
export const ComponentsNav = [];
export const FormsNav = [];
export const ChartsNav = [];
export const UiComponentsNav = [];
export const TablesNav = [];
export const DashboardsNav = [];
export const WidgetsNav = [];
export const AppsNav = [];
export const PagesNav = [];
export const UpgradeNav = [];