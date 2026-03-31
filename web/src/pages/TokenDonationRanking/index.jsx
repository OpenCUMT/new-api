import React, { useEffect, useMemo, useState } from 'react';
import {
    Button,
    Card,
    Space,
    Table,
    Tag,
    Typography,
} from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';
import { API, showError, timestamp2string } from '../../helpers';

const TokenDonationRankingPage = () => {
    const { t } = useTranslation();
    const [loading, setLoading] = useState(false);
    const [ranking, setRanking] = useState([]);

    const loadRanking = async () => {
        setLoading(true);
        try {
            const res = await API.get('/api/user/token_donation/ranking');
            const { success, message, data } = res.data;
            if (success) {
                const list = Array.isArray(data)
                    ? data.filter((item) => Number(item?.donation_count || 0) > 0)
                    : [];
                setRanking(list);
            } else {
                showError(message);
            }
        } catch (error) {
            showError(error.response?.data?.message || error.message);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        loadRanking();
    }, []);

    const stats = useMemo(() => {
        const contributorCount = ranking.length;
        const donationCount = ranking.reduce(
            (sum, item) => sum + Number(item.donation_count || 0),
            0,
        );
        return { contributorCount, donationCount };
    }, [ranking]);

    const columns = [
        {
            title: t('排名'),
            dataIndex: 'rank',
            width: 100,
            render: (value) => {
                if (value === 1) return <Tag color='red'>#{value}</Tag>;
                if (value === 2) return <Tag color='orange'>#{value}</Tag>;
                if (value === 3) return <Tag color='yellow'>#{value}</Tag>;
                return `#${value}`;
            },
        },
        {
            title: t('用户'),
            dataIndex: 'username',
            width: 220,
            render: (text) => text || '-',
        },
        {
            title: t('有效捐赠数'),
            dataIndex: 'donation_count',
            width: 160,
            render: (value) => Number(value || 0),
        },
        {
            title: t('最近有效捐赠时间'),
            dataIndex: 'latest_donation_time',
            width: 220,
            render: (value) => (value ? timestamp2string(value) : '-'),
        },
    ];

    return (
        <div className='w-full px-2 mt-[60px]'>
            <div className='mx-auto max-w-7xl space-y-4'>
                <Card className='!rounded-2xl'>
                    <div className='flex flex-col gap-3 md:flex-row md:items-end md:justify-between'>
                        <div>
                            <Typography.Title heading={4} className='!mb-1'>
                                {t('捐赠排行')}
                            </Typography.Title>
                            <Typography.Text type='secondary'>
                                {t('仅统计有效捐赠次数大于 0 的记录，按有效捐赠数从高到低排序。')}
                            </Typography.Text>
                        </div>
                        <Button loading={loading} onClick={loadRanking}>
                            {t('刷新排行')}
                        </Button>
                    </div>
                </Card>

                <div className='grid grid-cols-1 gap-4 md:grid-cols-2'>
                    <Card className='!rounded-2xl'>
                        <Space direction='vertical' spacing={4}>
                            <Typography.Text type='secondary'>
                                {t('参与有效捐赠人数')}
                            </Typography.Text>
                            <Typography.Title heading={3} className='!mb-0'>
                                {stats.contributorCount}
                            </Typography.Title>
                        </Space>
                    </Card>
                    <Card className='!rounded-2xl'>
                        <Space direction='vertical' spacing={4}>
                            <Typography.Text type='secondary'>
                                {t('有效捐赠总次数')}
                            </Typography.Text>
                            <Typography.Title heading={3} className='!mb-0'>
                                {stats.donationCount}
                            </Typography.Title>
                        </Space>
                    </Card>
                </div>

                <Card className='!rounded-2xl'>
                    <Table
                        rowKey='user_id'
                        loading={loading}
                        dataSource={ranking}
                        columns={columns}
                        pagination={false}
                        scroll={{ x: 700 }}
                    />
                </Card>
            </div>
        </div>
    );
};

export default TokenDonationRankingPage;
