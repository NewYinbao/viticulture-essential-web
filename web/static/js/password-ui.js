const $ = (s) => document.querySelector(s);
let seat = '',
  saving = false;

export function renderPassword(v) {
  const next = v.code + ':' + v.youId;
  if (seat !== next) {
    seat = next;
    $('#password-form').reset();
    $('#password-error').textContent = '';
    $('#password-settings').open = !v.passwordSet;
  }
  $('#password-settings-title').textContent = v.passwordSet
    ? '玩家密码 · 已保护'
    : '为旧座位设置密码';
  $('#current-password-label').hidden = !v.passwordSet;
  $('#current-password').required = !!v.passwordSet;
  $('#enrollment-code-label').hidden = !!v.passwordSet;
  $('#enrollment-code').required = !v.passwordSet;
  $('#password-settings-note').textContent = v.passwordSet
    ? '修改后，此座位在其他标签页或设备的登录会退出。'
    : '验证码仅显示在服务器本机窗口。设置密码后恢复对局。';
  $('#lock-session').disabled = !v.passwordSet;
  $('#lock-session').title = v.passwordSet
    ? '退出当前登录，保留对局座位'
    : '先设置密码，便于恢复座位';
}

export function clearPassword() {
  seat = '';
  $('#password-form').reset();
}

export function initPassword(api, changed) {
  $('#password-form').onsubmit = async (event) => {
    event.preventDefault();
    if (saving) return;
    const password = $('#new-password').value;
    if (password !== $('#confirm-password').value) {
      $('#password-error').textContent = '两次密码不一致';
      $('#confirm-password').focus();
      return;
    }
    saving = true;
    const button = $('#password-form button[type=submit]');
    button.disabled = true;
    $('#password-error').textContent = '';
    try {
      const result = await api('/api/password', {
        password,
        currentPassword: $('#current-password').value,
        enrollmentCode: $('#enrollment-code').value,
      });
      $('#password-form').reset();
      $('#password-settings').open = false;
      await changed(result.token);
    } catch (error) {
      $('#password-error').textContent = error.message;
    } finally {
      saving = false;
      button.disabled = false;
    }
  };
}
