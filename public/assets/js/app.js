$(function() {

    var targetSwitchEle = $('input[name="rm_target_dir"]');
    // Switches
    if (targetSwitchEle.length) {
      targetSwitchEle.bootstrapSwitch();
    }

    // 监听modal隐藏事件
    $('#delete_hook_modal').on('hide.bs.modal', function(e) {
        targetSwitchEle.bootstrapSwitch('state', false);
    });

    alertify.set({
        buttonReverse: true,
        labels: {
            ok: '是',
            cancel: '否'
        }
    });

    $('#for_new_repos').on('click', function(e) {
        e.preventDefault();

        $('#new_repos_modal').modal('show');
    });

    $('#for_new_hook').on('click', function(e) {
        e.preventDefault();

        $('#new_hook_modal').modal('show');
    });

    $('#button_new_repos').on('click', function(e) {
        e.preventDefault();

        var adapterID = $('#select_adapter > option:selected').val(),
            reposName = $('input[name="repos_name"]').val(),
            reposRemote = $('input[name="repos_remote"]').val();

        if (adapterID === '' || reposName === '') {
            alertify.log('类型和仓库名称均不能为空！', 'error', 5000);
            return;
        }

        var req = $.ajax({
            'type': 'post',
            'url': '/new/repos',
            'data': {
                'repos_type': adapterID,
                'repos_name': reposName,
                'repos_remote': reposRemote
            },
            'dataType': 'json'
        });
        req.done(function (resp) {
            $('#new_repos_modal').modal('hide');
            if (resp.Status === 'success') {
                alertify.log(resp.Msg, 'success', 1000);
                setTimeout("window.location.href='/'", 1500);
            } else {
                 alertify.log(resp.Msg, 'error', 5000);
            }
        });
    });

    $('#button_new_hook').on('click', function(e) {
        e.preventDefault();

        var targetRepos = $('#select_repos > option:selected').val(),
            branchName = $('input[name="branch_name"]').val(),
            targetDir = $('input[name="target_dir"]').val();

        if (targetRepos === '' || branchName === '' || targetDir === '') {
            alertify.log('三项均不能为空', 'error', 5000);
            return;
        }

        var req = $.ajax({
            'type': 'post',
            'url': '/new/hook',
            'data': {
                'repos_id': targetRepos,
                'which_branch': branchName,
                'target_dir': targetDir
            },
            'dataType': 'json'
        });
        req.done(function (resp) {
            $('#new_hook_modal').modal('hide');
            if (resp.Status === 'success') {
                alertify.log(resp.Msg, 'success', 1000);
                setTimeout("window.location.href='/'", 1500);
            } else {
                 alertify.log(resp.Msg, 'error', 5000);
            }
        });
    });

    $('a.delete-it').on('click', function(e) {
        e.preventDefault();
        e.stopPropagation();

        var myParent = $(this).parent(),
            hookID = myParent.siblings('.hook-id').text();

        $('input[name="hook_id_to_delete"]').val(hookID);

        $('#delete_hook_modal').modal('show');
    });

    $('a.modify-it').on('click', function(e) {
        e.preventDefault();
        e.stopPropagation();

        var myParent = $(this).parent(),
            hookID = myParent.siblings('.hook-id').text(),
            branchName = myParent.siblings('.branch-name').text(),
            targetDir = myParent.siblings('.target-dir').text();

        $('input[name="branch_name_update"]').val(branchName);
        $('input[name="target_dir_update"]').val(targetDir);
        $('input[name="hook_id_to_update"]').val(hookID);

        $('#update_hook_modal').modal('show');
    });

    $('#button_update_hook').on('click', function(e) {
        e.preventDefault();
        e.stopPropagation();

        var branchNameUpdate = $('input[name="branch_name_update"]').val(),
            targetDirUpdate = $('input[name="target_dir_update"]').val(),
            hookIDToUpdate = $('input[name="hook_id_to_update"]').val();

        if (branchNameUpdate === '' || targetDirUpdate === '') {
            alertify.log('分支名和目标目录不能为空', 'error', 5000);
            return;
        }

        var req = $.ajax({
            'type': 'post',
            'url': '/update/hook',
            'data': {
                hook_id: hookIDToUpdate,
                branch_name: branchNameUpdate,
                target_dir: targetDirUpdate
            },
            'dataType': 'json'
        });

        req.done(function (resp) {
            if (resp.Status === 'success') {
                alertify.log(resp.Msg, 'success', 1000);
                setTimeout("window.location.href='/'", 1500);
            } else {
                alertify.log(resp.Msg, 'error', 5000);
            }
        });
    });

    $('#button_delete_hook').on('click', function(e) {
        e.preventDefault();
        e.stopPropagation();

        var targetHookID = $('input[name="hook_id_to_delete"]').val(),
            alsoDeleteTargetDir = targetSwitchEle.prop('checked');

        $('#delete_hook_modal').modal('hide');

        var req = $.ajax({
            'type': 'post',
            'url': '/delete/hook',
            'data': {
                hook_id: targetHookID,
                erase_all: alsoDeleteTargetDir
            },
            'dataType': 'json'
        });

        req.done(function (resp) {
            if (resp.Status === 'success') {
                alertify.log(resp.Msg, 'success', 1000);
                setTimeout("window.location.href='/'", 1500);
            } else {
                alertify.log(resp.Msg, 'error', 5000);
            }
        });


    });

    $('.repos-title').on('dblclick', function(e) {
        e.preventDefault();
        e.stopPropagation();

        var targetElement = $(this).children('.panel-title').children('span'),
            reposID = targetElement.attr('title'),
            reposName = targetElement.text();
        alertify.confirm('你确定删除' + reposName +'仓库吗？', function (e) {
            if (e) {
                alertify.log('你选择了"是"', '', 2000);
                var req = $.ajax({
                    'type': 'post',
                    'url': '/delete/repos',
                    'data': {
                        repos_id: reposID
                    },
                    'dataType': 'json'
                });

                req.done(function (resp) {
                    if (resp.Status === 'success') {
                        alertify.log(resp.Msg, 'success', 1000);
                        setTimeout("window.location.href='/'", 1500);
                    } else {
                        alertify.log(resp.Msg, 'error', 5000);
                    }

                });

            } else {
                alertify.log('你选择了"否"', '', 2000);
            }
        });
    });
});
