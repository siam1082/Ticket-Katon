// The people shown in "Meet the team" on the home page and in the footer.
// Edit this file only. Everything on the page comes from here.
//
// photo: put the image file in frontend/public/team/ and write its name here
//        (for example '/team/hasib.jpg'). If the file is missing, initials are shown instead.
// links: leave a link as '' to hide it. Use full addresses starting with https://
//        (email is just the address, like 'name@example.com').

export const PROJECT = {
  heading: 'Built by two students',
  subheading: 'TicketKaton is our academic project. We designed it, built it and tested it together.',
};

export const TEAM = [
  {
    id: 'hasib',
    name: 'Mohammad Hasibur Rahman',
    // role: 'Booking, payments and admin',
    education: '2nd-year CSE student · Shahjalal University of Science and Technology (SUST), Sylhet',
    bio: 'Worked on the booking and payment flow, ticket cancellation and the admin portal',
    photo: '/team/hasib.jpeg',
    links: { github: 'https://github.com/Hasib-2005', linkedin: 'https://www.linkedin.com/in/mohammad-hasibur-rahman-4941aa371/?isSelfProfile=true', facebook: 'https://www.facebook.com/share/1JVa8xMF7e/', email: 'hasib4362005@gmail.com', website: 'will publish later' },
  },
  {
    id: 'siam',
    name: 'Toufiqul Hossain Siam', // TODO: write your teammate's full name
    // role: 'Seat locking and frontend', // TODO: check this matches what they built
    education: '2nd-year CSE student · Shahjalal University of Science and Technology (SUST), Sylhet',
    bio: 'Worked on the seat locking engine, the first version of the frontend and the database setup.', // TODO
    photo: '/team/teammate.jpeg',
    links: { github: 'https://github.com/siam1082', linkedin: 'https://www.linkedin.com/in/toufiqul-siam-212b33343/', facebook: 'https://www.facebook.com/share/1A3bvVutvJ/', email: 'thsiam360@gmail.com', website: 'will publish later' },
  },
];
