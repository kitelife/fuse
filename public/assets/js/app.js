$(function() {
    if ($('[data-toggle="select"]').length) {
        $('[data-toggle="select"]').select2();
    }

    $('#button_new_repos').on('click', function(e) {
        e.preventDefaults();

        var pluginID = $('#select_plugin > option:selected').val(),
            reposName = $('input[name="repos_name"]').val(),
            reposRemote = $('input[name="repos_remote"]').val();

        if (pluginID === '' || reposName === '') {
            alertify.log('类型和仓库名称均不能为空！', 'error', 5000);
            return;
        }

        var req = $.ajax({
            'type': 'post',
            'url': '/new/repos',
            'data': {
                'repos_type': pluginID,
                'repos_name': reposName,
                'repos_remote': reposRemote
            },
            'dataType': 'json'
        });
        req.done(function (resp) {
            $('#new_repos_modal').model('hide');
            if (resp.Status === 'success') {
                alertify.log(resp.Msg, 'success', 1000);
                setTimeout("window.location.href='/'", 1500);
            } else {
                 alertify.log(resp.Msg, 'error', 5000);
            }
        });
    });

    $('#button_new_hook').on('click', function(e) {
        e.preventDefaults();

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
            $('#new_hook_modal').model('hide');
            if (resp.Status === 'success') {
                alertify.log(resp.Msg, 'success', 1000);
                setTimeout("window.location.href='/'", 1500);
            } else {
                 alertify.log(resp.Msg, 'error', 5000);
            }
        });
    });
});
