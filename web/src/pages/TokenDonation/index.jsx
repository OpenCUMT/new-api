import React, { useEffect, useState } from 'react';
import {
  Button,
  Card,
  Input,
  Select,
  Space,
  Table,
  TextArea,
  Tabs,
  TabPane,
  Tag,
  Typography,
} from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';
import {
  API,
  isAdmin,
  showError,
  showSuccess,
  timestamp2string,
} from '../../helpers';

const OPENAI_DONATION_CHANNEL_TYPE = 1;

const defaultFormState = {
  type: OPENAI_DONATION_CHANNEL_TYPE,
  name: '',
  key: '',
  base_url: '',
  token_source: '',
  models: '',
  group: 'default',
  remark: '',
};

const statusMeta = {
  pending: { color: 'orange', text: '待审核' },
  approved: { color: 'green', text: '已入池' },
  rejected: { color: 'red', text: '已拒绝' },
};

const TokenDonationPage = () => {
  const { t } = useTranslation();
  const adminMode = isAdmin();
  const [formState, setFormState] = useState(defaultFormState);
  const [submitting, setSubmitting] = useState(false);
  const [selfLoading, setSelfLoading] = useState(false);
  const [adminLoading, setAdminLoading] = useState(false);
  const [reviewingId, setReviewingId] = useState('');
  const [adminStatus, setAdminStatus] = useState('pending');
  const [selfDonations, setSelfDonations] = useState([]);
  const [adminDonations, setAdminDonations] = useState([]);

  const handleFormChange = (field, value) => {
    setFormState((prev) => ({ ...prev, [field]: value }));
  };

  const loadSelfDonations = async () => {
    setSelfLoading(true);
    try {
      const res = await API.get('/api/user/token_donation');
      const { success, message, data } = res.data;
      if (success) {
        setSelfDonations(data || []);
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.response?.data?.message || error.message);
    } finally {
      setSelfLoading(false);
    }
  };

  const loadAdminDonations = async (status = adminStatus) => {
    if (!adminMode) return;
    setAdminLoading(true);
    try {
      const res = await API.get(`/api/token_donation/?status=${status}`);
      const { success, message, data } = res.data;
      if (success) {
        setAdminDonations(data || []);
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.response?.data?.message || error.message);
    } finally {
      setAdminLoading(false);
    }
  };

  useEffect(() => {
    loadSelfDonations();
  }, []);

  useEffect(() => {
    loadAdminDonations(adminStatus);
  }, [adminStatus]);

  const handleSubmitDonation = async () => {
    if (!formState.name.trim()) {
      showError(t('请输入渠道名称'));
      return;
    }
    if (!formState.key.trim()) {
      showError(t('请输入要捐赠的 Token'));
      return;
    }
    if (!formState.models.trim()) {
      showError(t('请输入可用模型，多个模型请用英文逗号分隔'));
      return;
    }
    if (!formState.group.trim()) {
      showError(t('请输入调用分组'));
      return;
    }
    if (!formState.base_url.trim()) {
      showError(t('请输入 Base URL'));
      return;
    }
    if (!formState.token_source.trim()) {
      showError(t('请输入 token 来源'));
      return;
    }
    setSubmitting(true);
    try {
      const payload = {
        ...formState,
        type: OPENAI_DONATION_CHANNEL_TYPE,
        name: formState.name.trim(),
        key: formState.key.trim(),
        base_url: formState.base_url.trim(),
        token_source: formState.token_source.trim(),
        models: formState.models.trim(),
        group: formState.group.trim(),
        remark: formState.remark.trim(),
      };
      const res = await API.post('/api/user/token_donation', payload);
      const { success, message } = res.data;
      if (success) {
        showSuccess(t('Token 捐赠已提交，等待管理员审核'));
        setFormState(defaultFormState);
        await loadSelfDonations();
        await loadAdminDonations(adminStatus);
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.response?.data?.message || error.message);
    } finally {
      setSubmitting(false);
    }
  };

  const handleReview = async (id, action) => {
    setReviewingId(`${action}-${id}`);
    try {
      const res = await API.post(`/api/token_donation/${id}/${action}`);
      const { success, message } = res.data;
      if (success) {
        showSuccess(
          action === 'approve'
            ? t('已审核通过并加入调用池')
            : t('已拒绝该捐赠'),
        );
        await Promise.all([
          loadSelfDonations(),
          loadAdminDonations(adminStatus),
        ]);
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.response?.data?.message || error.message);
    } finally {
      setReviewingId('');
    }
  };

  const renderStatus = (_, record) => {
    const meta = statusMeta[record.status] || statusMeta.pending;
    return <Tag color={meta.color}>{t(meta.text)}</Tag>;
  };

  const selfColumns = [
    {
      title: t('ID'),
      dataIndex: 'id',
      width: 80,
    },
    {
      title: t('渠道类型'),
      dataIndex: 'type_name',
      width: 140,
    },
    {
      title: t('渠道名称'),
      dataIndex: 'name',
      width: 180,
    },
    {
      title: t('token来源'),
      dataIndex: 'token_source',
      width: 180,
      render: (text) => text || '-',
    },
    {
      title: t('模型'),
      dataIndex: 'models',
      width: 220,
      render: (text) => (
        <Typography.Text ellipsis={{ showTooltip: true }}>
          {text}
        </Typography.Text>
      ),
    },
    {
      title: t('分组'),
      dataIndex: 'group',
      width: 120,
    },
    {
      title: t('状态'),
      dataIndex: 'status',
      width: 110,
      render: renderStatus,
    },
    {
      title: t('已入池渠道'),
      dataIndex: 'channel_id',
      width: 120,
      render: (value) => value || '-',
    },
    {
      title: t('提交时间'),
      dataIndex: 'created_time',
      width: 180,
      render: (value) => timestamp2string(value),
    },
  ];

  const adminColumns = [
    {
      title: t('ID'),
      dataIndex: 'id',
      width: 80,
    },
    {
      title: t('用户'),
      dataIndex: 'username',
      width: 150,
      render: (_, record) => (
        <div className='flex flex-col'>
          <span>{record.username || '-'}</span>
          <span className='text-xs text-gray-500'>{record.email || '-'}</span>
        </div>
      ),
    },
    {
      title: t('渠道类型'),
      dataIndex: 'type_name',
      width: 140,
    },
    {
      title: t('渠道名称'),
      dataIndex: 'name',
      width: 180,
    },
    {
      title: t('token来源'),
      dataIndex: 'token_source',
      width: 200,
      render: (text) => text || '-',
    },
    {
      title: t('模型'),
      dataIndex: 'models',
      width: 220,
      render: (text) => (
        <Typography.Text ellipsis={{ showTooltip: true }}>
          {text}
        </Typography.Text>
      ),
    },
    {
      title: t('Base URL'),
      dataIndex: 'base_url',
      width: 220,
      render: (text) => text || '-',
    },
    {
      title: t('状态'),
      dataIndex: 'status',
      width: 110,
      render: renderStatus,
    },
    {
      title: t('操作'),
      dataIndex: 'operate',
      width: 180,
      fixed: 'right',
      render: (_, record) => (
        <Space>
          <Button
            size='small'
            type='primary'
            theme='solid'
            disabled={record.status !== 'pending'}
            loading={reviewingId === `approve-${record.id}`}
            onClick={() => handleReview(record.id, 'approve')}
          >
            {t('通过')}
          </Button>
          <Button
            size='small'
            type='danger'
            theme='outline'
            disabled={record.status !== 'pending'}
            loading={reviewingId === `reject-${record.id}`}
            onClick={() => handleReview(record.id, 'reject')}
          >
            {t('拒绝')}
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <div className='mt-[60px] px-2'>
      <div className='mx-auto max-w-7xl space-y-4'>
        <Card className='!rounded-2xl'>
          <div className='flex flex-col gap-2 md:flex-row md:items-end md:justify-between'>
            <div>
              <Typography.Title heading={4} className='!mb-1'>
                {t('Token 捐赠')}
              </Typography.Title>
              <Typography.Text type='secondary'>
                {t(
                  '普通用户可以在这里提交上游 Token，管理员审核通过后会自动生成渠道并加入后台调用池。',
                )}
              </Typography.Text>
            </div>
            <div className='text-sm text-gray-500'>
              {t('除备注外其余字段均为必填，请务必准确填写渠道类型、渠道名称和 Base URL。')}
            </div>
          </div>
        </Card>

        <Tabs type='card' defaultActiveKey='submit'>
          <TabPane tab={t('提交捐赠')} itemKey='submit'>
            <Card className='!rounded-2xl'>
              <div className='grid grid-cols-1 gap-4 lg:grid-cols-2'>
                <div>
                  <div className='mb-2 text-sm font-medium'>
                    {t('渠道类型（必填）')}
                  </div>
                  <Input value='OpenAI' disabled />
                  <Typography.Text type='tertiary' className='mt-1 block text-xs'>
                    {t('固定为 OpenAI 类型，无需选择。')}
                  </Typography.Text>
                </div>
                <div>
                  <div className='mb-2 text-sm font-medium'>
                    {t('渠道名称（必填）')}
                  </div>
                  <Input
                    value={formState.name}
                    onChange={(value) => handleFormChange('name', value)}
                    placeholder={t('例如：xxx捐赠的GPT Plus')}
                  />
                  <Typography.Text type='tertiary' className='mt-1 block text-xs'>
                    {t('将在后台以该名称保存 token，例如：xxx捐赠的GPT Plus。')}
                  </Typography.Text>
                </div>
                <div className='lg:col-span-2'>
                  <div className='mb-2 text-sm font-medium'>
                    {t('Token（必填）')}
                  </div>
                  <TextArea
                    value={formState.key}
                    onChange={(value) => handleFormChange('key', value)}
                    autosize={{ minRows: 3, maxRows: 6 }}
                    placeholder={t('输入要捐赠的上游 Token / Key')}
                  />
                </div>
                <div>
                  <div className='mb-2 text-sm font-medium'>
                    {t('模型列表（必填）')}
                  </div>
                  <Input
                    value={formState.models}
                    onChange={(value) => handleFormChange('models', value)}
                    placeholder={t('例如：gpt-4o,gpt-4.1-mini')}
                  />
                </div>
                <div>
                  <div className='mb-2 text-sm font-medium'>
                    {t('调用分组（必填）')}
                  </div>
                  <Input
                    value={formState.group}
                    onChange={(value) => handleFormChange('group', value)}
                    placeholder='default'
                  />
                </div>
                <div>
                  <div className='mb-2 text-sm font-medium'>
                    {t('Base URL（必填）')}
                  </div>
                  <Input
                    value={formState.base_url}
                    onChange={(value) => handleFormChange('base_url', value)}
                    placeholder={t('填写 token 来源提供的调用地址')}
                  />
                  <Typography.Text type='tertiary' className='mt-1 block text-xs'>
                    {t('填写 token 来源提供的调用地址，如果有多个，请填写 OpenAI 类型地址。')}
                  </Typography.Text>
                </div>
                <div>
                  <div className='mb-2 text-sm font-medium'>
                    {t('token来源（必填）')}
                  </div>
                  <Input
                    value={formState.token_source}
                    onChange={(value) => handleFormChange('token_source', value)}
                    placeholder={t('例如：OpenAI GPT Plus')}
                  />
                  <Typography.Text type='tertiary' className='mt-1 block text-xs'>
                    {t('AI 服务提供方及套餐，如：OpenAI GPT Plus。')}
                  </Typography.Text>
                </div>
                <div className='lg:col-span-2'>
                  <div className='mb-2 text-sm font-medium'>{t('备注')}</div>
                  <TextArea
                    value={formState.remark}
                    onChange={(value) => handleFormChange('remark', value)}
                    autosize={{ minRows: 2, maxRows: 4 }}
                    placeholder={t('可选，说明有效期、剩余 token 量或其他细节')}
                  />
                </div>
              </div>
              <div className='mt-4 flex justify-end'>
                <Button
                  theme='solid'
                  type='primary'
                  loading={submitting}
                  onClick={handleSubmitDonation}
                >
                  {t('提交捐赠')}
                </Button>
              </div>
            </Card>
          </TabPane>

          <TabPane tab={t('我的捐赠记录')} itemKey='self'>
            <Card className='!rounded-2xl'>
              <Table
                rowKey='id'
                loading={selfLoading}
                dataSource={selfDonations}
                columns={selfColumns}
                pagination={false}
                scroll={{ x: 1330 }}
              />
            </Card>
          </TabPane>

          {adminMode && (
            <TabPane tab={t('管理员审核')} itemKey='admin'>
              <Card className='!rounded-2xl'>
                <div className='mb-4 flex justify-end'>
                  <Select
                    value={adminStatus}
                    style={{ width: 180 }}
                    optionList={[
                      { value: 'pending', label: t('仅看待审核') },
                      { value: 'approved', label: t('仅看已通过') },
                      { value: 'rejected', label: t('仅看已拒绝') },
                      { value: 'all', label: t('查看全部') },
                    ]}
                    onChange={(value) => setAdminStatus(value)}
                  />
                </div>
                <Table
                  rowKey='id'
                  loading={adminLoading}
                  dataSource={adminDonations}
                  columns={adminColumns}
                  pagination={false}
                  scroll={{ x: 1600 }}
                />
              </Card>
            </TabPane>
          )}
        </Tabs>
      </div>
    </div>
  );
};

export default TokenDonationPage;
