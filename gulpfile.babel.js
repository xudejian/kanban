// generated on 2015-07-17 using generator-gulp-webapp 1.0.1
import gulp from 'gulp';
import gulpLoadPlugins from 'gulp-load-plugins';
import browserSync from 'browser-sync';
import del from 'del';
import modRewrite from 'connect-modrewrite';
var exec = require('child_process').exec;
import webpack from 'webpack';
import webpackConfig from './webpack.config.js';

const $ = gulpLoadPlugins();
const reload = browserSync.reload;

gulp.task("webpack", (callback) => {
  var myConfig = Object.create(webpackConfig);
  webpack(
      myConfig
      , function(err, stats) {
        if (err) {
          console.log(err);
          this.end();
        }
        callback();
      });
});

gulp.task('styles', () => {
  return gulp.src('app/styles/*.css')
    .pipe($.sourcemaps.init())
    .pipe($.autoprefixer({browsers: ['last 1 version']}))
    .pipe($.sourcemaps.write())
    .pipe(gulp.dest('.tmp/styles'))
    .pipe(reload({stream: true}));
});

function lint(files, options) {
  return () => {
    return gulp.src(files)
      .pipe(reload({stream: true, once: true}))
      .pipe($.eslint(options))
      .pipe($.eslint.format())
      .pipe($.if(!browserSync.active, $.eslint.failAfterError()));
  };
}
const testLintOptions = {
  env: {
    mocha: true
  },
  globals: {
    assert: false,
    expect: false,
    should: false
  }
};

gulp.task('lint', lint('app/scripts/**/*.js'));
gulp.task('lint:test', lint('test/spec/**/*.js', testLintOptions));

gulp.task('html', ['styles', 'webpack'], () => {
  const assets = $.useref.assets({searchPath: ['.tmp', 'app', '.']});

  return gulp.src('app/*.html')
    .pipe(assets)
    .pipe($.if('*.js', $.uglify()))
    .pipe($.if('*.css', $.minifyCss({compatibility: '*'})))
    .pipe(assets.restore())
    .pipe($.useref())
    .pipe($.if('*.html', $.minifyHtml({conditionals: true, loose: true})))
    .pipe(gulp.dest('dist/static'));
});

gulp.task('images', () => {
  return gulp.src('app/images/**/*')
    .pipe($.if($.if.isFile, $.cache($.imagemin({
      progressive: true,
      interlaced: true,
      // don't remove IDs from SVGs, they are often used
      // as hooks for embedding and styling
      svgoPlugins: [{cleanupIDs: false}]
    }))
    .on('error', function (err) {
      console.log(err);
      this.end();
    })))
    .pipe(gulp.dest('dist/static/images'));
});

gulp.task('conf', () => {
  return gulp.src([
    'conf/*.conf',
  ], {
    dot: false
  }).pipe(gulp.dest('dist/conf'));
});

gulp.task('go', () => {
  exec('go build -o dist/bin/kanpan ./serve', function (err, stdout, stderr) {
    console.log(stdout);
    console.log(stderr);
    if (err) {
      console.log(err);
    }
  });
});

gulp.task('extras', ['go', 'conf'], () => {
  return gulp.src([
    'app/*.*',
    '!app/**/*.coffee',
    '!app/*.html'
  ], {
    dot: true
  }).pipe(gulp.dest('dist/static'));
});

gulp.task('clean', del.bind(null, ['.tmp', 'dist']));

gulp.task('serve', ['styles', 'webpack'], () => {
  browserSync({
    notify: false,
    port: 9000,
    server: {
      baseDir: ['.tmp', 'app'],
      middleware: [
      modRewrite(['^/(stock/.*)$ http://127.0.0.1:3002/$1 [P]'])
      ]
    }
  });

  gulp.watch([
    'app/*.html',
    'app/scripts/**/*.js',
    'app/images/**/*',
    '.tmp/scripts/**/*.js',
  ]).on('change', reload);

  gulp.watch('app/scripts/**/*.coffee', ['webpack']);
  gulp.watch('app/styles/**/*.css', ['styles']);
});

gulp.task('serve:dist', () => {
  browserSync({
    notify: false,
    port: 9000,
    server: {
      baseDir: ['dist/static']
    },
    middleware: [
    modRewrite(['^/(stock/.*)$ http://127.0.0.1:3002/$1 [P]'])
    ]
  });
});

gulp.task('serve:test', () => {
  browserSync({
    notify: false,
    port: 9000,
    ui: false,
    server: {
      baseDir: 'test',
      routes: {
        '/bower_components': 'bower_components'
      }
    }
  });

  gulp.watch('test/spec/**/*.js').on('change', reload);
  gulp.watch('test/spec/**/*.js', ['lint:test']);
});

gulp.task('build', ['lint', 'html', 'images', 'extras'], () => {
  return gulp.src('dist/**/*').pipe($.size({title: 'build', gzip: true}));
});

gulp.task('default', ['clean'], () => {
  gulp.start('build');
});
