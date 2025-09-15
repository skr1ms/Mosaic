export const MainNav = [
    {
        icon: 'pe-7s-home',
        label: 'navigation.dashboard',
        to: '/dashboard',
        translate: true,
    },
];

export const AdminNav = [
    {
        icon: 'pe-7s-users',
        label: 'partners.title',
        translate: true,
        content: [
            {
                label: 'partners.list',
                to: '/partners/list',
                translate: true,
            },
            {
                label: 'partners.create',
                to: '/partners/create',
                translate: true,
            },
            {
                label: 'partners.blocked',
                to: '/partners/blocked',
                translate: true,
            },
        ],
    },
    {
        icon: 'pe-7s-ticket',
        label: 'coupons.title',
        translate: true,
        content: [
            {
                label: 'coupons.create',
                to: '/coupons/create',
                translate: true,
            },
            {
                label: 'coupons.manage',
                to: '/coupons/manage',
                translate: true,
            },
            {
                label: 'coupons.activated',
                to: '/coupons/activated',
                translate: true,
            },
        ],
    },
    {
        icon: 'pe-7s-graph1',
        label: 'analytics.title',
        translate: true,
        content: [
            {
                label: 'analytics.overview',
                to: '/analytics',
                translate: true,
            },
            {
                label: 'analytics.partners',
                to: '/analytics/partners',
                translate: true,
            },
            {
                label: 'images.queue',
                to: '/images/queue',
                translate: true,
            },
        ],
    },
];

export const PartnerNav = [
    {
        icon: 'pe-7s-ticket',
        label: 'partner_dashboard.my_coupons',
        translate: true,
        content: [
            {
                label: 'partner_dashboard.coupon_list',
                to: '/partner/coupons',
                translate: true,
            },
            {
                label: 'partner_dashboard.export_coupons',
                to: '/partner/coupons/export',
                translate: true,
            },
        ],
    },
    {
        icon: 'pe-7s-user',
        label: 'user.profile',
        translate: true,
        content: [
            {
                label: 'user.my_profile',
                to: '/partner/profile',
                translate: true,
            },
        ],
    },
    {
        icon: 'pe-7s-graph1',
        label: 'analytics.title',
        translate: true,
        content: [
            {
                label: 'analytics.my_analytics',
                to: '/partner/analytics',
                translate: true,
            },
        ],
    },
];

export const SystemNav = [
    {
        icon: 'pe-7s-monitor',
        label: 'system.system_status',
        translate: true,
        
        to: window.location.hostname.includes('doyoupaint.com') ? '/grafana/' : 'http://localhost:3000',
        external: true,
    },
    {
        icon: 'pe-7s-cloud',
        label: 'system.s3_minio',
        translate: true,
        
        to: window.location.hostname.includes('doyoupaint.com') ? '/minio/' : 'http://localhost:9001',
        external: true,
    },
    {
        icon: 'pe-7s-users',
        label: 'system.admin_management',
        translate: true,
        to: '/system/admins',
    },
];


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